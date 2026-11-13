package service

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/go-pdf/fpdf"
	"github.com/shopspring/decimal"

	"github.com/webond/api/internal/db"
	"github.com/webond/api/internal/domain"
)

// pdfCompress controls PDF stream compression. Production keeps it on; tests
// turn it off so the text content can be asserted on.
var pdfCompress = true

// statementPageSize is how many rows are read per page while walking a user's
// history/ledger back to the requested month.
const statementPageSize = 1000

// MonthlyStatement renders a real PDF statement for the given month (yyyyMM,
// US-028): account header, cash opening/closing balance taken from the wallet
// ledger, a summary of the month's activity, and the transaction list.
//
// Text is deliberately English with asset codes (not Thai names): the standard
// PDF fonts have no Thai glyphs and fpdf does not shape Thai combining marks,
// so Thai text would be garbled. The account is identified by email and id.
func (s *HistoryService) MonthlyStatement(ctx context.Context, userID uint64, yyyyMM string) ([]byte, string, error) {
	start, err := time.ParseInLocation("200601", yyyyMM, Bangkok)
	if len(yyyyMM) != 6 || err != nil {
		return nil, "", domain.NewError(domain.CodeValidationError, "Month must be in yyyyMM format")
	}
	now := s.clock.Now()
	currentMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, Bangkok)
	if start.After(currentMonth) {
		return nil, "", domain.NewError(domain.CodeValidationError, "A statement for a future month is not available")
	}
	end := start.AddDate(0, 1, 0)

	user, err := s.store.Users.GetByID(ctx, userID)
	if err != nil {
		return nil, "", domain.NewError(domain.CodeUnauthorized, "User not found")
	}

	rows, err := s.monthTransactions(ctx, userID, start, end)
	if err != nil {
		return nil, "", err
	}
	if len(rows) > maxExportRows {
		return nil, "", domain.NewError(domain.CodeExportTooLarge, "Too many transactions in this month for a statement")
	}
	opening, closing, err := s.cashBalances(ctx, userID, start, end)
	if err != nil {
		return nil, "", err
	}

	data, err := s.renderStatement(ctx, statementData{
		Email: user.Email, UserID: userID, Month: start, Generated: now,
		Opening: opening, Closing: closing, Rows: rows,
	})
	if err != nil {
		return nil, "", err
	}
	return data, fmt.Sprintf("webond-statement-%s.pdf", yyyyMM), nil
}

// monthTransactions walks the user's history newest-first until it has passed
// the start of the month, returning the month's rows oldest-first.
func (s *HistoryService) monthTransactions(ctx context.Context, userID uint64, start, end time.Time) ([]db.Transaction, error) {
	var inMonth []db.Transaction
	for offset := 0; ; offset += statementPageSize {
		page, err := s.store.Transactions.ListByUser(ctx, db.ListTransactionsByUserParams{
			UserID: userID, Limit: statementPageSize, Offset: int32(offset),
		})
		if err != nil {
			return nil, fmt.Errorf("list transactions: %w", err)
		}
		for _, r := range page {
			if !r.OccurredAt.Before(start) && r.OccurredAt.Before(end) {
				inMonth = append(inMonth, r)
			}
		}
		if len(page) < statementPageSize || page[len(page)-1].OccurredAt.Before(start) {
			break
		}
	}
	sort.SliceStable(inMonth, func(i, j int) bool { return inMonth[i].OccurredAt.Before(inMonth[j].OccurredAt) })
	return inMonth, nil
}

// cashBalances reads the wallet ledger (newest first) to find the cash balance
// at the start and at the end of the month: the balance_after of the latest
// entry before each boundary. A boundary with no earlier entry is 0.
func (s *HistoryService) cashBalances(ctx context.Context, userID uint64, start, end time.Time) (opening, closing decimal.Decimal, err error) {
	w, err := s.store.Wallets.GetByUser(ctx, userID)
	if err != nil {
		if isNoRows(err) {
			return decimal.Zero, decimal.Zero, nil
		}
		return decimal.Zero, decimal.Zero, fmt.Errorf("get wallet: %w", err)
	}
	gotClosing, gotOpening := false, false
	for offset := 0; !gotOpening; offset += statementPageSize {
		page, lerr := s.store.Wallets.ListTransactions(ctx, db.ListWalletTransactionsParams{
			WalletID: w.ID, Limit: statementPageSize, Offset: int32(offset),
		})
		if lerr != nil {
			return decimal.Zero, decimal.Zero, fmt.Errorf("list wallet ledger: %w", lerr)
		}
		for _, e := range page {
			if !gotClosing && e.CreatedAt.Before(end) {
				closing, gotClosing = e.BalanceAfter, true
			}
			if !gotOpening && e.CreatedAt.Before(start) {
				opening, gotOpening = e.BalanceAfter, true
				break
			}
		}
		if len(page) < statementPageSize {
			break
		}
	}
	return opening, closing, nil
}

// statementData is everything the renderer needs.
type statementData struct {
	Email     string
	UserID    uint64
	Month     time.Time
	Generated time.Time
	Opening   decimal.Decimal
	Closing   decimal.Decimal
	Rows      []db.Transaction
}

// statementTotals are completed-activity totals shown in the summary box.
type statementTotals struct {
	deposits, withdrawals, buys, sells, income decimal.Decimal
}

func totalsOf(rows []db.Transaction) statementTotals {
	var t statementTotals
	for _, r := range rows {
		if r.Status != db.TransactionsStatusCompleted {
			continue
		}
		amt := r.Amount.Abs()
		switch r.Type {
		case db.TransactionsTypeDeposit:
			t.deposits = t.deposits.Add(amt)
		case db.TransactionsTypeWithdraw:
			t.withdrawals = t.withdrawals.Add(amt)
		case db.TransactionsTypeBuy, db.TransactionsTypeDcaBuy:
			t.buys = t.buys.Add(amt)
		case db.TransactionsTypeSell:
			t.sells = t.sells.Add(amt)
		case db.TransactionsTypeCoupon, db.TransactionsTypeMaturity:
			t.income = t.income.Add(amt)
		}
	}
	return t
}

// Table geometry (mm) — the widths sum to the printable width of an A4 page.
var stmtCols = []struct {
	title string
	w     float64
	align string
}{
	{"Date", 27, "L"}, {"Reference", 33, "L"}, {"Type", 22, "L"}, {"Asset", 27, "L"},
	{"Quantity", 25, "R"}, {"Amount (THB)", 28, "R"}, {"Status", 18, "L"},
}

func (s *HistoryService) renderStatement(ctx context.Context, d statementData) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetCompression(pdfCompress)
	pdf.SetCreationDate(d.Generated)
	pdf.SetTitle("WeBond monthly statement "+d.Month.Format("2006-01"), false)
	pdf.SetAuthor("WeBond", false)
	pdf.SetMargins(15, 15, 15)
	pdf.SetAutoPageBreak(true, 18)
	pdf.AliasNbPages("")

	pdf.SetFooterFunc(func() {
		pdf.SetY(-12)
		pdf.SetFont("Helvetica", "I", 8)
		pdf.SetTextColor(110, 110, 110)
		pdf.CellFormat(0, 6, fmt.Sprintf("Page %d of {nb}", pdf.PageNo()), "", 0, "C", false, 0, "")
	})

	pdf.AddPage()

	// Brand + title.
	pdf.SetFont("Helvetica", "B", 20)
	pdf.SetTextColor(15, 37, 82)
	pdf.CellFormat(0, 10, "WeBond", "", 1, "L", false, 0, "")
	pdf.SetFont("Helvetica", "B", 13)
	pdf.SetTextColor(0, 0, 0)
	pdf.CellFormat(0, 7, "Monthly Statement - "+d.Month.Format("January 2006"), "", 1, "L", false, 0, "")
	pdf.Ln(2)

	// Account block.
	pdf.SetFont("Helvetica", "", 10)
	pdf.CellFormat(0, 5.5, latin("Account: "+d.Email+" (ID "+fmt.Sprint(d.UserID)+")"), "", 1, "L", false, 0, "")
	pdf.CellFormat(0, 5.5, "Period: "+d.Month.Format("02 Jan 2006")+" to "+d.Month.AddDate(0, 1, -1).Format("02 Jan 2006"), "", 1, "L", false, 0, "")
	pdf.CellFormat(0, 5.5, "Generated: "+d.Generated.Format("02 Jan 2006 15:04")+" (Asia/Bangkok)", "", 1, "L", false, 0, "")
	pdf.Ln(4)

	// Summary.
	t := totalsOf(d.Rows)
	pdf.SetFont("Helvetica", "B", 11)
	pdf.CellFormat(0, 7, "Cash summary (THB)", "B", 1, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 10)
	summary := []struct{ label, value string }{
		{"Opening cash balance", moneyText(d.Opening)},
		{"Deposits", moneyText(t.deposits)},
		{"Withdrawals", moneyText(t.withdrawals)},
		{"Purchases (Buy / DCA Buy)", moneyText(t.buys)},
		{"Sales", moneyText(t.sells)},
		{"Coupons and maturities", moneyText(t.income)},
		{"Closing cash balance", moneyText(d.Closing)},
	}
	for i, row := range summary {
		bold := i == 0 || i == len(summary)-1
		if bold {
			pdf.SetFont("Helvetica", "B", 10)
		} else {
			pdf.SetFont("Helvetica", "", 10)
		}
		pdf.CellFormat(110, 6, row.label, "", 0, "L", false, 0, "")
		pdf.CellFormat(70, 6, row.value, "", 1, "R", false, 0, "")
	}
	pdf.Ln(5)

	// Transactions.
	pdf.SetFont("Helvetica", "B", 11)
	pdf.CellFormat(0, 7, fmt.Sprintf("Transactions (%d)", len(d.Rows)), "", 1, "L", false, 0, "")

	const rowH = 6.0
	pageBottom := 297.0 - 18.0
	header := func() {
		pdf.SetFont("Helvetica", "B", 9)
		pdf.SetFillColor(15, 37, 82)
		pdf.SetTextColor(255, 255, 255)
		for _, c := range stmtCols {
			pdf.CellFormat(c.w, rowH, c.title, "", 0, c.align, true, 0, "")
		}
		pdf.Ln(-1)
		pdf.SetTextColor(0, 0, 0)
	}

	if len(d.Rows) == 0 {
		pdf.SetFont("Helvetica", "I", 10)
		pdf.CellFormat(0, 8, "No transactions in this month.", "", 1, "L", false, 0, "")
	} else {
		header()
		codes := map[string]string{}
		pdf.SetFont("Helvetica", "", 8.5)
		for i, r := range d.Rows {
			if pdf.GetY()+rowH > pageBottom {
				pdf.AddPage()
				header()
				pdf.SetFont("Helvetica", "", 8.5)
			}
			if i%2 == 1 {
				pdf.SetFillColor(244, 246, 250)
			} else {
				pdf.SetFillColor(255, 255, 255)
			}
			qty := ""
			if r.Quantity.Valid {
				qty = r.Quantity.Decimal.StringFixed(4)
			}
			cells := []string{
				r.OccurredAt.In(Bangkok).Format("02 Jan 15:04"),
				latin(r.TxnRef),
				latin(string(r.Type)),
				latin(s.assetCode(ctx, codes, r)),
				qty,
				moneyText(r.Amount),
				latin(string(r.Status)),
			}
			for ci, c := range stmtCols {
				pdf.CellFormat(c.w, rowH, cells[ci], "", 0, c.align, true, 0, "")
			}
			pdf.Ln(-1)
		}
	}

	pdf.Ln(6)
	pdf.SetFont("Helvetica", "I", 8)
	pdf.SetTextColor(110, 110, 110)
	pdf.MultiCell(0, 4.2, "This statement is generated from demonstration data for the WeBond sample application. "+
		"It is not a legal, tax or regulated financial document.", "", "L", false)

	if err := pdf.Error(); err != nil {
		return nil, fmt.Errorf("render statement: %w", err)
	}
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("write statement pdf: %w", err)
	}
	return buf.Bytes(), nil
}

// assetCode resolves a transaction's asset to its short code (cached per call);
// cash transactions (no asset) show a dash.
func (s *HistoryService) assetCode(ctx context.Context, cache map[string]string, r db.Transaction) string {
	if !r.AssetType.Valid || !r.AssetID.Valid {
		return "-"
	}
	key := fmt.Sprintf("%s:%d", r.AssetType.TransactionsAssetType, r.AssetID.Int64)
	if code, ok := cache[key]; ok {
		return code
	}
	id := uint64(r.AssetID.Int64)
	code := "-"
	switch r.AssetType.TransactionsAssetType {
	case db.TransactionsAssetTypeFund:
		if f, err := s.store.Funds.GetByID(ctx, id); err == nil {
			code = f.Code
		}
	case db.TransactionsAssetTypeBond:
		if b, err := s.store.Bonds.GetByID(ctx, id); err == nil {
			code = b.Code
		}
	case db.TransactionsAssetTypeGold:
		if g, err := s.store.Gold.GetByID(ctx, id); err == nil {
			code = g.Code
		}
	}
	cache[key] = code
	return code
}

// moneyText formats a THB amount with thousands separators and 2 decimals,
// e.g. -1234567.8 -> "-1,234,567.80".
func moneyText(d decimal.Decimal) string {
	s := d.StringFixed(2)
	sign := ""
	if strings.HasPrefix(s, "-") {
		sign, s = "-", s[1:]
	}
	intPart, frac := s, ""
	if i := strings.IndexByte(s, '.'); i >= 0 {
		intPart, frac = s[:i], s[i:]
	}
	var b strings.Builder
	for i, r := range intPart {
		if i > 0 && (len(intPart)-i)%3 == 0 {
			b.WriteByte(',')
		}
		b.WriteRune(r)
	}
	return sign + b.String() + frac
}

// latin replaces characters the standard PDF fonts cannot draw (anything
// outside printable ASCII) with '?', so a stray non-Latin character never
// produces a corrupt glyph.
func latin(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= 32 && r < 127 {
			b.WriteRune(r)
		} else {
			b.WriteByte('?')
		}
	}
	return b.String()
}
