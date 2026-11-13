package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/shopspring/decimal"

	"github.com/webond/api/internal/db"
	"github.com/webond/api/internal/domain"
	"github.com/webond/api/internal/money"
	"github.com/webond/api/internal/repo"
)

// quoteTTL is the gold price-lock window (EX-08, US-017).
const quoteTTL = 30 * time.Second

// mfCutoff is the daily cut-off hour/minute for same-day NAV (BR-02, 15:30).
var mfCutoffHour, mfCutoffMinute = 15, 30

// TradingService implements Buy/Sell for MF, Bond and Gold with all BR/EX
// rules, idempotency, the gold quote lock, risk-mismatch ack, holdings avg-cost
// updates, and the transaction + status-history writes — everything for one
// order inside a single DB transaction (components.md § TradingService;
// correctness.md P3,P5; NFR-05).
type TradingService struct {
	store  *repo.Store
	wallet *WalletService
	notify *NotificationService
	clock  Clock

	quotes *quoteStore
}

// NewTradingService builds a TradingService.
func NewTradingService(store *repo.Store, wallet *WalletService, notify *NotificationService, clock Clock) *TradingService {
	return &TradingService{store: store, wallet: wallet, notify: notify, clock: clock, quotes: newQuoteStore()}
}

// OrderInput is the place-order request (api-spec.md POST /api/orders).
type OrderInput struct {
	UserID         uint64
	AssetType      string // fund|bond|gold
	AssetID        uint64
	Side           string // buy|sell
	OrderMode      string // amount|units|weight
	Amount         decimal.Decimal
	Units          decimal.Decimal
	WeightGrams    decimal.Decimal
	RiskAck        bool
	IdempotencyKey string
	// internal: DCA linkage (set by DCAService)
	dcaPlanID  int64
	dcaRoundID int64
	dcaType    bool
}

// OrderResult is the created/settled order view returned to the client.
type OrderResult struct {
	TxnRef     string `json:"txnRef"`
	Status     string `json:"status"`
	AssetType  string `json:"assetType"`
	AssetID    uint64 `json:"assetId"`
	Side       string `json:"side"`
	Amount     string `json:"amount,omitempty"`
	Units      string `json:"units,omitempty"`
	Weight     string `json:"weightGrams,omitempty"`
	Price      string `json:"price,omitempty"`
	NetAmount  string `json:"netAmount,omitempty"`
	RealizedPL string `json:"realizedPl,omitempty"`
	TradeDate  string `json:"tradeDate,omitempty"`
	SettleDate string `json:"settlementDate,omitempty"`
}

// PlaceOrder routes to the asset-specific flow after shared pre-checks
// (suitability BR-01, risk-mismatch ack BR-08/EX-12, idempotency EX-11).
func (s *TradingService) PlaceOrder(ctx context.Context, in OrderInput) (OrderResult, error) {
	if !assetTypeValid(in.AssetType) {
		return OrderResult{}, domain.NewError(domain.CodeValidationError, "Unknown asset type")
	}
	if in.Side != "buy" && in.Side != "sell" {
		return OrderResult{}, domain.NewError(domain.CodeValidationError, "Side must be buy or sell")
	}

	// Idempotency (EX-11): return the original order if the key was seen.
	if in.IdempotencyKey != "" {
		if existing, err := s.store.Orders.GetByIdempotencyKey(ctx, nullString(in.IdempotencyKey)); err == nil {
			return s.orderResultFromRow(existing), nil
		} else if !isNoRows(err) {
			return OrderResult{}, fmt.Errorf("idempotency lookup: %w", err)
		}
	}

	// Suitability: account active + valid risk assessment (BR-01).
	if err := s.ensureTradeable(ctx, in.UserID); err != nil {
		return OrderResult{}, err
	}

	switch in.AssetType {
	case "fund":
		return s.tradeFund(ctx, in)
	case "bond":
		return s.tradeBond(ctx, in)
	case "gold":
		return s.tradeGold(ctx, in)
	default:
		return OrderResult{}, domain.NewError(domain.CodeValidationError, "Unknown asset type")
	}
}

// ensureTradeable enforces BR-01: user active + non-expired risk assessment.
func (s *TradingService) ensureTradeable(ctx context.Context, userID uint64) error {
	u, err := s.store.Users.GetByID(ctx, userID)
	if err != nil {
		return domain.NewError(domain.CodeUnauthorized, "User not found")
	}
	if u.Status != db.UsersStatusActive {
		return domain.NewError(domain.CodeForbidden, "Please verify your email before trading")
	}
	r, err := s.store.Risk.GetLatest(ctx, userID)
	if err != nil {
		if isNoRows(err) {
			return domain.NewError(domain.CodeSuitabilityRequired, "Please complete the risk suitability assessment")
		}
		return fmt.Errorf("get risk: %w", err)
	}
	if s.clock.Now().After(r.ExpiresAt) {
		return domain.NewError(domain.CodeSuitabilityRequired, "Your risk assessment has expired; please retake it")
	}
	return nil
}

// requireRiskAck enforces BR-08/EX-12: if the asset risk level exceeds the
// user's level, an explicit acknowledgement is required; the ack is audited.
func (s *TradingService) requireRiskAck(ctx context.Context, userID uint64, assetRisk int16, ack bool) error {
	if assetRisk <= 0 {
		return nil
	}
	r, err := s.store.Risk.GetLatest(ctx, userID)
	if err != nil {
		return nil // suitability already checked; absence handled upstream
	}
	if int16(r.RiskLevel) >= assetRisk {
		return nil
	}
	if !ack {
		return domain.NewError(domain.CodeRiskAckRequired, "This asset exceeds your risk level; acknowledgement required")
	}
	s.audit(ctx, userID, "risk_ack", "order", 0)
	return nil
}

// --- Mutual Fund (US-013/014, BR-02/03, EX-04/05/06/11) ---

func (s *TradingService) tradeFund(ctx context.Context, in OrderInput) (OrderResult, error) {
	fund, err := s.store.Funds.GetByID(ctx, in.AssetID)
	if err != nil {
		return OrderResult{}, assetNotFound(err)
	}
	nav, err := s.store.Funds.GetLatestNav(ctx, fund.ID)
	if err != nil {
		return OrderResult{}, domain.NewError(domain.CodePriceUnavailable, "Fund NAV is unavailable")
	}
	if fund.RiskLevel.Valid {
		if rerr := s.requireRiskAck(ctx, in.UserID, fund.RiskLevel.Int16, in.RiskAck); rerr != nil {
			return OrderResult{}, rerr
		}
	}

	tradeDate, settleDate := s.navDates(fund)

	var out OrderResult
	err = s.store.WithinTx(ctx, func(tx *repo.Store) error {
		if in.Side == "buy" {
			amount := money.Quantize(in.Amount)
			if amount.LessThan(fund.MinPurchase) {
				return domain.NewError(domain.CodeValidationError,
					fmt.Sprintf("Minimum purchase is %s THB", money.FormatTHB(fund.MinPurchase)))
			}
			units, uerr := money.AllotUnits(amount, nav.Nav)
			if uerr != nil {
				return domain.NewError(domain.CodePriceUnavailable, "Invalid NAV for allotment")
			}
			res, oerr := s.createAndSettleBuy(ctx, tx, in, db.OrdersAssetTypeFund, db.OrdersOrderModeAmount,
				amount, units, decimal.Zero, nav.Nav, tradeDate, settleDate, db.HoldingsAssetTypeFund, money.TxnBuy)
			out = res
			return oerr
		}
		// Sell by units or by THB amount (US-014): verify holding, credit proceeds at NAV.
		units, uerr := sellUnits(in, nav.Nav)
		if uerr != nil {
			return uerr
		}
		res, serr := s.sellPosition(ctx, tx, in, db.OrdersAssetTypeFund, db.HoldingsAssetTypeFund,
			units, nav.Nav, tradeDate, settleDate, money.TxnSell)
		out = res
		return serr
	})
	return out, err
}

// navDates computes trade/settlement dates for an MF order respecting the 15:30
// cut-off (BR-02): after cut-off → next business day NAV; settlement is T+n.
func (s *TradingService) navDates(fund db.Fund) (time.Time, time.Time) {
	now := s.clock.Now()
	trade := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, Bangkok)
	cutoff := time.Date(now.Year(), now.Month(), now.Day(), mfCutoffHour, mfCutoffMinute, 0, 0, Bangkok)
	if now.After(cutoff) {
		trade = nextBusinessDay(trade.AddDate(0, 0, 1))
	} else {
		trade = nextBusinessDay(trade)
	}
	settle := addBusinessDays(trade, int(fund.SettlementDays))
	return trade, settle
}

// nextBusinessDay skips weekends (phase-1 approximation; holidays not modelled).
func nextBusinessDay(t time.Time) time.Time {
	for t.Weekday() == time.Saturday || t.Weekday() == time.Sunday {
		t = t.AddDate(0, 0, 1)
	}
	return t
}

// addBusinessDays adds n business days skipping weekends.
func addBusinessDays(t time.Time, n int) time.Time {
	for i := 0; i < n; i++ {
		t = t.AddDate(0, 0, 1)
		t = nextBusinessDay(t)
	}
	return t
}

// --- Bond (US-015/016, BR-04→Add-on, EX-07) ---

func (s *TradingService) tradeBond(ctx context.Context, in OrderInput) (OrderResult, error) {
	bond, err := s.store.Bonds.GetByID(ctx, in.AssetID)
	if err != nil {
		return OrderResult{}, assetNotFound(err)
	}
	if bond.RiskLevel.Valid {
		if rerr := s.requireRiskAck(ctx, in.UserID, bond.RiskLevel.Int16, in.RiskAck); rerr != nil {
			return OrderResult{}, rerr
		}
	}
	now := s.clock.Now()

	var out OrderResult
	err = s.store.WithinTx(ctx, func(tx *repo.Store) error {
		if in.Side == "buy" {
			// Offering window (EX-07).
			if bond.OfferingStart.Valid && now.Before(bond.OfferingStart.Time) {
				return domain.NewError(domain.CodeBondOfferInvalid, "Bond offering has not opened")
			}
			if bond.OfferingEnd.Valid && now.After(bond.OfferingEnd.Time.AddDate(0, 0, 1)) {
				return domain.NewError(domain.CodeBondOfferInvalid, "Bond offering is closed")
			}
			amount := money.Quantize(in.Amount)
			// Lot rule: ≥ 100,000 and exact multiple (Add-on, correctness.md P5).
			if amount.LessThan(bond.MinLot) {
				return domain.NewError(domain.CodeBondOfferInvalid,
					fmt.Sprintf("Minimum lot is %s THB", money.FormatTHB(bond.MinLot)))
			}
			if lerr := money.ValidateBondLot(amount, bond.MinLot); lerr != nil {
				return domain.NewError(domain.CodeBondOfferInvalid,
					fmt.Sprintf("Amount must be a multiple of %s THB", money.FormatTHB(bond.MinLot)))
			}
			// Units = face-value units = amount / faceValue (BR-05, P5).
			units, uerr := money.AllotUnits(amount, bond.FaceValue)
			if uerr != nil {
				return domain.NewError(domain.CodeBondOfferInvalid, "Invalid face value")
			}
			tradeDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, Bangkok)
			res, oerr := s.createAndSettleBuy(ctx, tx, in, db.OrdersAssetTypeBond, db.OrdersOrderModeAmount,
				amount, units, decimal.Zero, bond.FaceValue, tradeDate, tradeDate, db.HoldingsAssetTypeBond, money.TxnBuy)
			out = res
			return oerr
		}
		// Sell bond at clean price + accrued (US-016).
		price, perr := s.store.Bonds.GetLatestPrice(ctx, bond.ID)
		if perr != nil {
			return domain.NewError(domain.CodePriceUnavailable, "Bond price is unavailable")
		}
		// Dirty price per unit in THB: clean price is quoted as % of par, so convert
		// it with the face value, then add accrued interest (already THB per unit).
		unitPrice := money.Quantize(money.BondUnitPrice(bond.FaceValue, price.CleanPrice).Add(price.AccruedInterest))
		units, uerr := sellUnits(in, unitPrice)
		if uerr != nil {
			return uerr
		}
		tradeDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, Bangkok)
		res, serr := s.sellPosition(ctx, tx, in, db.OrdersAssetTypeBond, db.HoldingsAssetTypeBond,
			units, unitPrice, tradeDate, tradeDate, money.TxnSell)
		out = res
		return serr
	})
	return out, err
}

// --- Gold (US-017, BR-09, EX-08/10) ---

// GoldQuote is a 30-second locked price quote (US-017).
type GoldQuote struct {
	QuoteID   string `json:"quoteId"`
	Price     string `json:"price"`
	Weight    string `json:"weight"`
	Total     string `json:"total"`
	ExpiresAt string `json:"expiresAt"`
}

// Quote locks the current gold price for 30s for a buy/sell (US-017). Blocks if
// the price is stale (>5 min, EX-10).
func (s *TradingService) Quote(ctx context.Context, userID, goldID uint64, side string, amount, weightGrams decimal.Decimal) (GoldQuote, error) {
	if side != "buy" && side != "sell" {
		return GoldQuote{}, domain.NewError(domain.CodeValidationError, "Side must be buy or sell")
	}
	price, err := s.store.Gold.GetLatestPrice(ctx, goldID)
	if err != nil {
		return GoldQuote{}, domain.NewError(domain.CodePriceUnavailable, "Gold price is unavailable")
	}
	if s.clock.Now().Sub(price.PricedAt) > goldStaleAfter {
		return GoldQuote{}, domain.NewError(domain.CodePriceUnavailable, "Gold price is unavailable (stale)")
	}
	goldSide := money.GoldBuy
	if side == "sell" {
		goldSide = money.GoldSell
	}
	unit := money.GoldUnitPrice(goldSide, price.BidPerGram, price.AskPerGram)

	var grams, total decimal.Decimal
	if weightGrams.GreaterThan(decimal.Zero) {
		grams = money.Quantize(weightGrams)
		total = money.GoldTradeValue(goldSide, grams, price.BidPerGram, price.AskPerGram)
	} else {
		if amount.LessThanOrEqual(decimal.Zero) {
			return GoldQuote{}, domain.NewError(domain.CodeValidationError, "Provide an amount or weight")
		}
		grams = money.Quantize(amount.DivRound(unit, money.Scale+8))
		total = money.GoldTradeValue(goldSide, grams, price.BidPerGram, price.AskPerGram)
	}

	q := s.quotes.put(quote{
		userID: userID, goldID: goldID, side: side,
		unit: unit, grams: grams, total: total,
		expiresAt: s.clock.Now().Add(quoteTTL),
	})
	return GoldQuote{
		QuoteID: q.id, Price: money.String(unit), Weight: money.FormatQuantity(grams),
		Total: money.FormatTHB(total), ExpiresAt: q.expiresAt.Format("2006-01-02T15:04:05-07:00"),
	}, nil
}

// ConfirmQuote executes a locked gold quote within 30s (US-017). Expired → 409
// QUOTE_EXPIRED (EX-08).
func (s *TradingService) ConfirmQuote(ctx context.Context, userID uint64, quoteID string, idempotencyKey string) (OrderResult, error) {
	q, ok := s.quotes.get(quoteID)
	if !ok || q.userID != userID {
		return OrderResult{}, domain.NewError(domain.CodeQuoteExpired, "Price quote expired, please re-quote")
	}
	if s.clock.Now().After(q.expiresAt) {
		s.quotes.del(quoteID)
		return OrderResult{}, domain.NewError(domain.CodeQuoteExpired, "Price quote expired, please re-quote")
	}
	if err := s.ensureTradeable(ctx, userID); err != nil {
		return OrderResult{}, err
	}
	s.quotes.del(quoteID)

	in := OrderInput{
		UserID: userID, AssetType: "gold", AssetID: q.goldID, Side: q.side,
		OrderMode: "weight", WeightGrams: q.grams, IdempotencyKey: idempotencyKey,
	}
	var out OrderResult
	err := s.store.WithinTx(ctx, func(tx *repo.Store) error {
		if q.side == "buy" {
			res, oerr := s.createAndSettleBuy(ctx, tx, in, db.OrdersAssetTypeGold, db.OrdersOrderModeWeight,
				q.total, q.grams, q.grams, q.unit, s.today(), s.today(), db.HoldingsAssetTypeGold, money.TxnBuy)
			out = res
			return oerr
		}
		res, serr := s.sellGold(ctx, tx, in, q.grams, q.unit)
		out = res
		return serr
	})
	return out, err
}

// tradeGold handles a direct gold order (no explicit quote) by quoting and
// confirming in one step at the current price (used by DCA buys of gold).
func (s *TradingService) tradeGold(ctx context.Context, in OrderInput) (OrderResult, error) {
	price, err := s.store.Gold.GetLatestPrice(ctx, in.AssetID)
	if err != nil {
		return OrderResult{}, domain.NewError(domain.CodePriceUnavailable, "Gold price is unavailable")
	}
	if s.clock.Now().Sub(price.PricedAt) > goldStaleAfter {
		return OrderResult{}, domain.NewError(domain.CodePriceUnavailable, "Gold price is unavailable (stale)")
	}
	goldSide := money.GoldBuy
	if in.Side == "sell" {
		goldSide = money.GoldSell
	}
	unit := money.GoldUnitPrice(goldSide, price.BidPerGram, price.AskPerGram)

	var grams, total decimal.Decimal
	if in.WeightGrams.GreaterThan(decimal.Zero) {
		grams = money.Quantize(in.WeightGrams)
	} else {
		grams = money.Quantize(in.Amount.DivRound(unit, money.Scale+8))
	}
	total = money.GoldTradeValue(goldSide, grams, price.BidPerGram, price.AskPerGram)

	var out OrderResult
	err = s.store.WithinTx(ctx, func(tx *repo.Store) error {
		if in.Side == "buy" {
			res, oerr := s.createAndSettleBuy(ctx, tx, in, db.OrdersAssetTypeGold, db.OrdersOrderModeWeight,
				total, grams, grams, unit, s.today(), s.today(), db.HoldingsAssetTypeGold, money.TxnBuy)
			out = res
			return oerr
		}
		res, serr := s.sellGold(ctx, tx, in, grams, unit)
		out = res
		return serr
	})
	return out, err
}

func (s *TradingService) sellGold(ctx context.Context, tx *repo.Store, in OrderInput, grams, unit decimal.Decimal) (OrderResult, error) {
	return s.sellPosition(ctx, tx, in, db.OrdersAssetTypeGold, db.HoldingsAssetTypeGold, grams, unit, s.today(), s.today(), money.TxnSell)
}

func (s *TradingService) today() time.Time {
	now := s.clock.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, Bangkok)
}

// createAndSettleBuy performs a buy atomically: reserve+settle cash, update the
// holding's average cost (BR-10), create the order (completed) and write the
// transaction + status history. Everything runs in the passed tx (NFR-05, P3).
func (s *TradingService) createAndSettleBuy(
	ctx context.Context, tx *repo.Store, in OrderInput,
	assetType db.OrdersAssetType, mode db.OrdersOrderMode,
	amount, qty, weight, unitPrice decimal.Decimal,
	tradeDate, settleDate time.Time, holdType db.HoldingsAssetType, txnType money.TxnType,
) (OrderResult, error) {
	now := s.clock.Now()
	txnRef := NewTxnRef(now)

	// Reserve then settle the cash (available must cover the buy).
	if err := s.wallet.ReserveTx(ctx, tx, in.UserID, amount, 0); err != nil {
		return OrderResult{}, err
	}
	if err := s.wallet.SettleBuyTx(ctx, tx, in.UserID, amount, 0); err != nil {
		return OrderResult{}, err
	}

	// Average-cost update (BR-10). qty is units (fund/bond) or grams (gold).
	if err := s.applyBuyHolding(ctx, tx, in.UserID, holdType, in.AssetID, qty, unitPrice); err != nil {
		return OrderResult{}, err
	}

	orderID, err := s.insertOrder(ctx, tx, in, txnRef, assetType, mode, db.OrdersSideBuy,
		amount, qty, weight, unitPrice, amount, decimal.NullDecimal{}, db.OrdersStatusCompleted, tradeDate, settleDate)
	if err != nil {
		return OrderResult{}, err
	}

	txnID, err := s.insertTransaction(ctx, tx, in.UserID, txnRef, txnType, assetType, in.AssetID,
		money.SignedAmount(txnType, amount), qty, unitPrice, amount, db.TransactionsStatusCompleted,
		orderID, tradeDate, settleDate, decimal.NullDecimal{})
	if err != nil {
		return OrderResult{}, err
	}
	if err := s.appendStatus(ctx, tx, txnID, "submitted", now); err != nil {
		return OrderResult{}, err
	}
	if err := s.appendStatus(ctx, tx, txnID, "completed", now); err != nil {
		return OrderResult{}, err
	}

	return OrderResult{
		TxnRef: txnRef, Status: "completed", AssetType: string(assetType), AssetID: in.AssetID,
		Side: "buy", Amount: money.FormatTHB(amount), Units: money.FormatQuantity(qty),
		Price: money.String(unitPrice), NetAmount: money.FormatTHB(amount),
		TradeDate: tradeDate.Format("2006-01-02"), SettleDate: settleDate.Format("2006-01-02"),
	}, nil
}

// sellPosition performs a sell atomically: verify holding, reduce it (avg-cost,
// realized P/L BR-10/P3), credit proceeds, create order + transaction + status.
func (s *TradingService) sellPosition(
	ctx context.Context, tx *repo.Store, in OrderInput,
	assetType db.OrdersAssetType, holdType db.HoldingsAssetType,
	qty, unitPrice decimal.Decimal, tradeDate, settleDate time.Time, txnType money.TxnType,
) (OrderResult, error) {
	if qty.LessThanOrEqual(decimal.Zero) {
		return OrderResult{}, domain.NewError(domain.CodeValidationError, "Sell quantity must be positive")
	}
	now := s.clock.Now()
	txnRef := NewTxnRef(now)

	h, err := tx.Holdings.GetForUpdate(ctx, db.GetHoldingForUpdateParams{
		UserID: in.UserID, AssetType: holdType, AssetID: in.AssetID,
	})
	if err != nil {
		if isNoRows(err) {
			return OrderResult{}, domain.NewError(domain.CodeInsufficientHolding, "You do not hold this asset")
		}
		return OrderResult{}, fmt.Errorf("lock holding: %w", err)
	}
	available := h.Quantity.Sub(h.LockedQuantity)
	if qty.GreaterThan(available) {
		return OrderResult{}, domain.NewError(domain.CodeInsufficientHolding,
			fmt.Sprintf("You can sell up to %s units (about %s THB)",
				available.StringFixed(money.Scale), money.FormatTHB(money.Quantize(available.Mul(unitPrice)))))
	}

	res, serr := money.ApplySell(h.Quantity, h.TotalCost, qty, unitPrice)
	if serr != nil {
		return OrderResult{}, domain.NewError(domain.CodeInsufficientHolding, serr.Error())
	}
	proceeds := money.Quantize(qty.Mul(unitPrice))

	if err := tx.Holdings.Upsert(ctx, db.UpsertHoldingParams{
		UserID: in.UserID, AssetType: holdType, AssetID: in.AssetID,
		Quantity: res.Quantity, LockedQuantity: h.LockedQuantity,
		AvgCost: res.AvgCost, TotalCost: res.TotalCost,
	}); err != nil {
		return OrderResult{}, fmt.Errorf("update holding: %w", err)
	}

	// Credit sell proceeds to the wallet (settlement).
	if err := s.wallet.CreditTx(ctx, tx, in.UserID, proceeds, db.WalletTransactionsTypeSettlement, 0); err != nil {
		return OrderResult{}, err
	}

	orderID, err := s.insertOrder(ctx, tx, in, txnRef, assetType, db.OrdersOrderModeUnits, db.OrdersSideSell,
		proceeds, qty, qty, unitPrice, proceeds, nullDecimal(res.RealizUS), db.OrdersStatusCompleted, tradeDate, settleDate)
	if err != nil {
		return OrderResult{}, err
	}
	txnID, err := s.insertTransaction(ctx, tx, in.UserID, txnRef, txnType, assetType, in.AssetID,
		money.SignedAmount(txnType, proceeds), qty, unitPrice, proceeds, db.TransactionsStatusCompleted,
		orderID, tradeDate, settleDate, nullDecimal(res.RealizUS))
	if err != nil {
		return OrderResult{}, err
	}
	if err := s.appendStatus(ctx, tx, txnID, "submitted", now); err != nil {
		return OrderResult{}, err
	}
	if err := s.appendStatus(ctx, tx, txnID, "completed", now); err != nil {
		return OrderResult{}, err
	}

	return OrderResult{
		TxnRef: txnRef, Status: "completed", AssetType: string(assetType), AssetID: in.AssetID,
		Side: "sell", Units: money.FormatQuantity(qty), Price: money.String(unitPrice),
		NetAmount: money.FormatTHB(proceeds), RealizedPL: money.String(res.RealizUS),
		TradeDate: tradeDate.Format("2006-01-02"), SettleDate: settleDate.Format("2006-01-02"),
	}, nil
}

// applyBuyHolding loads (or creates) the holding and applies a buy at avg cost.
func (s *TradingService) applyBuyHolding(ctx context.Context, tx *repo.Store, userID uint64, holdType db.HoldingsAssetType, assetID uint64, qty, unitPrice decimal.Decimal) error {
	h, err := tx.Holdings.GetForUpdate(ctx, db.GetHoldingForUpdateParams{
		UserID: userID, AssetType: holdType, AssetID: assetID,
	})
	var curQty, curCost decimal.Decimal
	if err == nil {
		curQty, curCost = h.Quantity, h.TotalCost
	} else if !isNoRows(err) {
		return fmt.Errorf("lock holding: %w", err)
	}
	res, berr := money.ApplyBuy(curQty, curCost, qty, unitPrice)
	if berr != nil {
		return domain.NewError(domain.CodeValidationError, berr.Error())
	}
	return tx.Holdings.Upsert(ctx, db.UpsertHoldingParams{
		UserID: userID, AssetType: holdType, AssetID: assetID,
		Quantity: res.Quantity, LockedQuantity: decimal.Zero, AvgCost: res.AvgCost, TotalCost: res.TotalCost,
	})
}

// insertOrder writes an order row and returns its id.
func (s *TradingService) insertOrder(
	ctx context.Context, tx *repo.Store, in OrderInput, txnRef string,
	assetType db.OrdersAssetType, mode db.OrdersOrderMode, side db.OrdersSide,
	amount, units, weight, price, netAmount decimal.Decimal, realized decimal.NullDecimal,
	status db.OrdersStatus, tradeDate, settleDate time.Time,
) (uint64, error) {
	params := db.CreateOrderParams{
		TxnRef: txnRef, UserID: in.UserID, AssetType: assetType, AssetID: in.AssetID,
		Side: side, OrderMode: mode,
		Amount: nullDecimal(amount), Units: nullDecimal(units),
		Price: nullDecimal(price), Fee: decimal.Zero, NetAmount: nullDecimal(netAmount),
		Status: status, RealizedPl: realized, RiskAck: in.RiskAck, Channel: "web",
		TradeDate: nullTime(tradeDate), SettlementDate: nullTime(settleDate),
	}
	if assetType == db.OrdersAssetTypeGold {
		params.WeightGrams = nullDecimal(weight)
	}
	if in.IdempotencyKey != "" {
		params.IdempotencyKey = nullString(in.IdempotencyKey)
	}
	if in.dcaType {
		params.DcaPlanID = nullInt64(in.dcaPlanID)
		params.DcaRoundID = nullInt64(in.dcaRoundID)
	}
	return lastInsertID(tx.Orders.Create(ctx, params))
}

// insertTransaction writes a transaction row and returns its id.
func (s *TradingService) insertTransaction(
	ctx context.Context, tx *repo.Store, userID uint64, txnRef string, txnType money.TxnType,
	assetType db.OrdersAssetType, assetID uint64,
	signedAmount, qty, price, netAmount decimal.Decimal, status db.TransactionsStatus,
	orderID uint64, tradeDate, settleDate time.Time, realized decimal.NullDecimal,
) (uint64, error) {
	t := txTypeFor(txnType, assetType)
	params := db.CreateTransactionParams{
		TxnRef: txnRef, UserID: userID, Type: t,
		AssetType: db.NullTransactionsAssetType{TransactionsAssetType: db.TransactionsAssetType(assetType), Valid: true},
		AssetID:   nullInt64(int64(assetID)),
		Amount:    money.Quantize(signedAmount), Quantity: nullDecimal(qty), Price: nullDecimal(price),
		Fee: decimal.Zero, NetAmount: money.Quantize(netAmount), Status: status,
		OrderID: nullInt64(int64(orderID)), RealizedPl: realized,
		TradeDate: nullTime(tradeDate), SettlementDate: nullTime(settleDate),
		OccurredAt: s.clock.Now(),
	}
	return lastInsertID(tx.Transactions.Create(ctx, params))
}

// txTypeFor maps a money.TxnType to the DB transaction type.
func txTypeFor(t money.TxnType, _ db.OrdersAssetType) db.TransactionsType {
	switch t {
	case money.TxnBuy:
		return db.TransactionsTypeBuy
	case money.TxnSell:
		return db.TransactionsTypeSell
	case money.TxnDcaBuy:
		return db.TransactionsTypeDcaBuy
	default:
		return db.TransactionsTypeBuy
	}
}

// appendStatus appends a status-history row (FR-18, BR-14).
func (s *TradingService) appendStatus(ctx context.Context, tx *repo.Store, txnID uint64, status string, at time.Time) error {
	_, err := tx.Transactions.CreateStatusHistory(ctx, db.CreateTransactionStatusHistoryParams{
		TransactionID: txnID, Status: status, ChangedBy: nullString("system"), ChangedAt: at,
	})
	return err
}

// GetOrder returns an order by txn_ref enforcing ownership (US-026, EX-15).
func (s *TradingService) GetOrder(ctx context.Context, userID uint64, txnRef string) (OrderResult, error) {
	o, err := s.store.Orders.GetByTxnRef(ctx, txnRef)
	if err != nil || o.UserID != userID {
		return OrderResult{}, domain.NewError(domain.CodeTxnNotFound, "Transaction not found")
	}
	return s.orderResultFromRow(o), nil
}

func (s *TradingService) orderResultFromRow(o db.Order) OrderResult {
	r := OrderResult{
		TxnRef: o.TxnRef, Status: string(o.Status), AssetType: string(o.AssetType),
		AssetID: o.AssetID, Side: string(o.Side),
	}
	if o.Amount.Valid {
		r.Amount = money.FormatTHB(o.Amount.Decimal)
	}
	if o.Units.Valid {
		r.Units = money.FormatQuantity(o.Units.Decimal)
	}
	if o.WeightGrams.Valid {
		r.Weight = money.FormatQuantity(o.WeightGrams.Decimal)
	}
	if o.Price.Valid {
		r.Price = money.String(o.Price.Decimal)
	}
	if o.NetAmount.Valid {
		r.NetAmount = money.FormatTHB(o.NetAmount.Decimal)
	}
	if o.RealizedPl.Valid {
		r.RealizedPL = money.String(o.RealizedPl.Decimal)
	}
	if o.TradeDate.Valid {
		r.TradeDate = o.TradeDate.Time.Format("2006-01-02")
	}
	if o.SettlementDate.Valid {
		r.SettleDate = o.SettlementDate.Time.Format("2006-01-02")
	}
	return r
}

func (s *TradingService) audit(ctx context.Context, userID uint64, action, entity string, entityID int64) {
	_, _ = s.store.Audit.Create(ctx, db.CreateAuditLogParams{
		UserID: nullInt64(int64(userID)), Action: action, Entity: nullString(entity),
		EntityID: nullInt64(entityID),
	})
}

// --- in-memory gold quote store (30s TTL) ---

type quote struct {
	id        string
	userID    uint64
	goldID    uint64
	side      string
	unit      decimal.Decimal
	grams     decimal.Decimal
	total     decimal.Decimal
	expiresAt time.Time
}

type quoteStore struct {
	mu  sync.Mutex
	m   map[string]quote
	seq uint64
}

func newQuoteStore() *quoteStore { return &quoteStore{m: make(map[string]quote)} }

func (q *quoteStore) put(v quote) quote {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.seq++
	v.id = fmt.Sprintf("q_%d_%d", time.Now().UnixNano(), q.seq)
	q.m[v.id] = v
	return v
}

func (q *quoteStore) get(id string) (quote, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	v, ok := q.m[id]
	return v, ok
}

func (q *quoteStore) del(id string) {
	q.mu.Lock()
	defer q.mu.Unlock()
	delete(q.m, id)
}

// sellUnits turns a fund/bond sell request into units (US-014/016): by units
// (orderMode "units") or by a THB amount converted at the current unit price
// (orderMode "amount", or units left empty).
func sellUnits(in OrderInput, unitPrice decimal.Decimal) (decimal.Decimal, error) {
	if in.OrderMode != "amount" && in.Units.GreaterThan(decimal.Zero) {
		return money.Quantize(in.Units), nil
	}
	if in.Amount.LessThanOrEqual(decimal.Zero) {
		return decimal.Zero, domain.NewError(domain.CodeValidationError, "Enter an amount or a number of units to sell")
	}
	units, err := money.AllotUnits(in.Amount, unitPrice)
	if err != nil {
		return decimal.Zero, domain.NewError(domain.CodePriceUnavailable, "Price is unavailable")
	}
	return units, nil
}
