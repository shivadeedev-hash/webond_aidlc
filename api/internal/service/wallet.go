package service

import (
	"context"
	"fmt"

	"github.com/shopspring/decimal"

	"github.com/webond/api/internal/db"
	"github.com/webond/api/internal/domain"
	"github.com/webond/api/internal/money"
	"github.com/webond/api/internal/repo"
)

// WalletService implements cash deposit/withdraw and the reserve/release/settle
// primitives used by trading (components.md § WalletService; FR-07, BR-11).
// All mutations run inside Store.WithinTx with a row lock so the cash invariants
// hold under concurrency (correctness.md P2).
type WalletService struct {
	store *repo.Store
	clock Clock
}

// NewWalletService builds a WalletService.
func NewWalletService(store *repo.Store, clock Clock) *WalletService {
	return &WalletService{store: store, clock: clock}
}

// Balances is the wallet view returned to the client (US-010).
type Balances struct {
	Balance   decimal.Decimal
	Reserved  decimal.Decimal
	Available decimal.Decimal
}

// Get returns the user's balances with available = balance − reserved (US-010).
func (s *WalletService) Get(ctx context.Context, userID uint64) (Balances, error) {
	w, err := s.store.Wallets.GetByUser(ctx, userID)
	if err != nil {
		if isNoRows(err) {
			return Balances{}, domain.NewError(domain.CodeValidationError, "Wallet not found")
		}
		return Balances{}, fmt.Errorf("get wallet: %w", err)
	}
	return balancesOf(w), nil
}

// balancesOf computes the Balances view from a wallet row.
func balancesOf(w db.Wallet) Balances {
	return Balances{
		Balance:   w.Balance,
		Reserved:  w.Reserved,
		Available: money.Quantize(w.Balance.Sub(w.Reserved)),
	}
}

// Deposit adds cash to the wallet and records a +Deposit ledger row (US-008,
// BR-15). Atomic (row-locked) transaction.
func (s *WalletService) Deposit(ctx context.Context, userID uint64, amount decimal.Decimal) (Balances, error) {
	amount = money.Quantize(amount)
	if amount.LessThanOrEqual(decimal.Zero) {
		return Balances{}, domain.NewError(domain.CodeValidationError, "Deposit amount must be positive")
	}
	var out Balances
	err := s.store.WithinTx(ctx, func(tx *repo.Store) error {
		w, err := tx.Wallets.GetByUserForUpdate(ctx, userID)
		if err != nil {
			return fmt.Errorf("lock wallet: %w", err)
		}
		newBalance := money.Quantize(w.Balance.Add(amount))
		if err := tx.Wallets.UpdateBalances(ctx, db.UpdateWalletBalancesParams{
			Balance: newBalance, Reserved: w.Reserved, ID: w.ID,
		}); err != nil {
			return fmt.Errorf("update balances: %w", err)
		}
		txnID, err := s.recordCashTransaction(ctx, tx, userID, money.TxnDeposit, db.TransactionsTypeDeposit, amount)
		if err != nil {
			return err
		}
		if _, err := tx.Wallets.CreateTransaction(ctx, db.CreateWalletTransactionParams{
			WalletID:     w.ID,
			Type:         db.WalletTransactionsTypeDeposit,
			Amount:       money.SignedAmount(money.TxnDeposit, amount),
			BalanceAfter: newBalance,
			RefTxnID:     nullInt64(int64(txnID)),
			Status:       db.WalletTransactionsStatusDone,
		}); err != nil {
			return fmt.Errorf("ledger deposit: %w", err)
		}
		out = Balances{Balance: newBalance, Reserved: w.Reserved, Available: money.Quantize(newBalance.Sub(w.Reserved))}
		return nil
	})
	return out, err
}

// Withdraw removes available cash and records a −Withdraw ledger row (US-009,
// BR-11, BR-15). Rejects if amount exceeds available (INSUFFICIENT_BALANCE).
func (s *WalletService) Withdraw(ctx context.Context, userID uint64, amount decimal.Decimal) (Balances, error) {
	amount = money.Quantize(amount)
	if amount.LessThanOrEqual(decimal.Zero) {
		return Balances{}, domain.NewError(domain.CodeValidationError, "Withdrawal amount must be positive")
	}
	var out Balances
	err := s.store.WithinTx(ctx, func(tx *repo.Store) error {
		w, err := tx.Wallets.GetByUserForUpdate(ctx, userID)
		if err != nil {
			return fmt.Errorf("lock wallet: %w", err)
		}
		available := w.Balance.Sub(w.Reserved)
		if amount.GreaterThan(available) {
			return domain.NewError(domain.CodeInsufficientBalance, "Insufficient available balance")
		}
		newBalance := money.Quantize(w.Balance.Sub(amount))
		if err := tx.Wallets.UpdateBalances(ctx, db.UpdateWalletBalancesParams{
			Balance: newBalance, Reserved: w.Reserved, ID: w.ID,
		}); err != nil {
			return fmt.Errorf("update balances: %w", err)
		}
		txnID, err := s.recordCashTransaction(ctx, tx, userID, money.TxnWithdraw, db.TransactionsTypeWithdraw, amount)
		if err != nil {
			return err
		}
		if _, err := tx.Wallets.CreateTransaction(ctx, db.CreateWalletTransactionParams{
			WalletID:     w.ID,
			Type:         db.WalletTransactionsTypeWithdraw,
			Amount:       money.SignedAmount(money.TxnWithdraw, amount),
			BalanceAfter: newBalance,
			RefTxnID:     nullInt64(int64(txnID)),
			Status:       db.WalletTransactionsStatusDone,
		}); err != nil {
			return fmt.Errorf("ledger withdraw: %w", err)
		}
		out = Balances{Balance: newBalance, Reserved: w.Reserved, Available: money.Quantize(newBalance.Sub(w.Reserved))}
		return nil
	})
	return out, err
}

// recordCashTransaction writes the History-feed row for a completed cash
// movement (US-024 lists Deposit and Withdraw alongside trades), with a
// reference number and a status timeline entry. It runs in the caller's
// transaction so the feed and the wallet ledger can never disagree.
func (s *WalletService) recordCashTransaction(
	ctx context.Context, tx *repo.Store, userID uint64,
	kind money.TxnType, txType db.TransactionsType, amount decimal.Decimal,
) (uint64, error) {
	now := s.clock.Now()
	txnID, err := lastInsertID(tx.Transactions.Create(ctx, db.CreateTransactionParams{
		TxnRef: NewTxnRef(now), UserID: userID, Type: txType,
		Amount: money.SignedAmount(kind, amount), NetAmount: money.Quantize(amount),
		Fee: decimal.Zero, Status: db.TransactionsStatusCompleted, OccurredAt: now,
	}))
	if err != nil {
		return 0, fmt.Errorf("record %s transaction: %w", txType, err)
	}
	if _, err := tx.Transactions.CreateStatusHistory(ctx, db.CreateTransactionStatusHistoryParams{
		TransactionID: txnID, Status: "completed", ChangedBy: nullString("system"), ChangedAt: now,
	}); err != nil {
		return 0, fmt.Errorf("record %s status: %w", txType, err)
	}
	return txnID, nil
}

// ListTransactions returns a page of the wallet ledger (US-008/009/010).
func (s *WalletService) ListTransactions(ctx context.Context, userID uint64, page, pageSize int) ([]db.WalletTransaction, int64, error) {
	w, err := s.store.Wallets.GetByUser(ctx, userID)
	if err != nil {
		if isNoRows(err) {
			return nil, 0, domain.NewError(domain.CodeValidationError, "Wallet not found")
		}
		return nil, 0, fmt.Errorf("get wallet: %w", err)
	}
	limit, offset := pageBounds(page, pageSize)
	rows, err := s.store.Wallets.ListTransactions(ctx, db.ListWalletTransactionsParams{
		WalletID: w.ID, Limit: int32(limit), Offset: int32(offset),
	})
	if err != nil {
		return nil, 0, fmt.Errorf("list wallet txns: %w", err)
	}
	total, err := s.store.Wallets.CountTransactions(ctx, w.ID)
	if err != nil {
		return nil, 0, fmt.Errorf("count wallet txns: %w", err)
	}
	return rows, total, nil
}

// --- Reserve/release/settle primitives (BR-11) ---
//
// These operate on a transaction-bound store (tx) because they always run as
// part of a larger order mutation (reserve on pending buy, release on cancel,
// settle on completion). They enforce the cash invariants (correctness.md P2):
//   reserved ≥ 0, balance ≥ 0, reserved ≤ balance, available ≥ 0.

// ReserveTx moves `amount` from available into reserved (hold for a pending
// buy). Rejects if available < amount (INSUFFICIENT_BALANCE).
func (s *WalletService) ReserveTx(ctx context.Context, tx *repo.Store, userID uint64, amount decimal.Decimal, refTxnID int64) error {
	amount = money.Quantize(amount)
	if amount.LessThanOrEqual(decimal.Zero) {
		return domain.NewError(domain.CodeValidationError, "Reserve amount must be positive")
	}
	w, err := tx.Wallets.GetByUserForUpdate(ctx, userID)
	if err != nil {
		return fmt.Errorf("lock wallet: %w", err)
	}
	available := w.Balance.Sub(w.Reserved)
	if amount.GreaterThan(available) {
		return domain.NewError(domain.CodeInsufficientBalance, "Insufficient available balance")
	}
	newReserved := money.Quantize(w.Reserved.Add(amount))
	if err := tx.Wallets.UpdateBalances(ctx, db.UpdateWalletBalancesParams{
		Balance: w.Balance, Reserved: newReserved, ID: w.ID,
	}); err != nil {
		return fmt.Errorf("reserve: %w", err)
	}
	return s.ledgerHold(ctx, tx, w.ID, db.WalletTransactionsTypeHold, amount.Neg(), w.Balance, refTxnID)
}

// ReleaseTx returns `amount` from reserved back to available (order cancelled),
// leaving balance unchanged. Rejects releasing more than reserved.
func (s *WalletService) ReleaseTx(ctx context.Context, tx *repo.Store, userID uint64, amount decimal.Decimal, refTxnID int64) error {
	amount = money.Quantize(amount)
	if amount.LessThanOrEqual(decimal.Zero) {
		return domain.NewError(domain.CodeValidationError, "Release amount must be positive")
	}
	w, err := tx.Wallets.GetByUserForUpdate(ctx, userID)
	if err != nil {
		return fmt.Errorf("lock wallet: %w", err)
	}
	if amount.GreaterThan(w.Reserved) {
		return domain.NewError(domain.CodeValidationError, "Cannot release more than reserved")
	}
	newReserved := money.Quantize(w.Reserved.Sub(amount))
	if err := tx.Wallets.UpdateBalances(ctx, db.UpdateWalletBalancesParams{
		Balance: w.Balance, Reserved: newReserved, ID: w.ID,
	}); err != nil {
		return fmt.Errorf("release: %w", err)
	}
	return s.ledgerHold(ctx, tx, w.ID, db.WalletTransactionsTypeRelease, amount, w.Balance, refTxnID)
}

// SettleBuyTx finalises a reserved buy: it debits `amount` from both balance and
// reserved (the held cash leaves the wallet). Rejects settling more than
// reserved/balance.
func (s *WalletService) SettleBuyTx(ctx context.Context, tx *repo.Store, userID uint64, amount decimal.Decimal, refTxnID int64) error {
	amount = money.Quantize(amount)
	if amount.LessThanOrEqual(decimal.Zero) {
		return domain.NewError(domain.CodeValidationError, "Settle amount must be positive")
	}
	w, err := tx.Wallets.GetByUserForUpdate(ctx, userID)
	if err != nil {
		return fmt.Errorf("lock wallet: %w", err)
	}
	if amount.GreaterThan(w.Reserved) || amount.GreaterThan(w.Balance) {
		return domain.NewError(domain.CodeInsufficientBalance, "Reserved cash is insufficient to settle")
	}
	newBalance := money.Quantize(w.Balance.Sub(amount))
	newReserved := money.Quantize(w.Reserved.Sub(amount))
	if err := tx.Wallets.UpdateBalances(ctx, db.UpdateWalletBalancesParams{
		Balance: newBalance, Reserved: newReserved, ID: w.ID,
	}); err != nil {
		return fmt.Errorf("settle buy: %w", err)
	}
	return s.ledgerHold(ctx, tx, w.ID, db.WalletTransactionsTypeSettlement, amount.Neg(), newBalance, refTxnID)
}

// CreditTx adds `amount` to the balance (sell proceeds, coupon, maturity) with a
// typed ledger row (used by trading/bond lifecycle). Positive credit.
func (s *WalletService) CreditTx(ctx context.Context, tx *repo.Store, userID uint64, amount decimal.Decimal, kind db.WalletTransactionsType, refTxnID int64) error {
	amount = money.Quantize(amount)
	if amount.LessThanOrEqual(decimal.Zero) {
		return domain.NewError(domain.CodeValidationError, "Credit amount must be positive")
	}
	w, err := tx.Wallets.GetByUserForUpdate(ctx, userID)
	if err != nil {
		return fmt.Errorf("lock wallet: %w", err)
	}
	newBalance := money.Quantize(w.Balance.Add(amount))
	if err := tx.Wallets.UpdateBalances(ctx, db.UpdateWalletBalancesParams{
		Balance: newBalance, Reserved: w.Reserved, ID: w.ID,
	}); err != nil {
		return fmt.Errorf("credit: %w", err)
	}
	return s.ledgerHold(ctx, tx, w.ID, kind, amount, newBalance, refTxnID)
}

// ledgerHold appends a wallet_transactions row. signedAmount is already signed.
func (s *WalletService) ledgerHold(ctx context.Context, tx *repo.Store, walletID uint64, kind db.WalletTransactionsType, signedAmount, balanceAfter decimal.Decimal, refTxnID int64) error {
	var ref = nullInt64(refTxnID)
	if refTxnID == 0 {
		ref.Valid = false
	}
	if _, err := tx.Wallets.CreateTransaction(ctx, db.CreateWalletTransactionParams{
		WalletID:     walletID,
		Type:         kind,
		Amount:       money.Quantize(signedAmount),
		BalanceAfter: money.Quantize(balanceAfter),
		RefTxnID:     ref,
		Status:       db.WalletTransactionsStatusDone,
	}); err != nil {
		return fmt.Errorf("ledger %s: %w", kind, err)
	}
	return nil
}
