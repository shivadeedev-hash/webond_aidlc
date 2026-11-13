package service

import (
	"context"
	"fmt"
	"time"

	"github.com/shopspring/decimal"

	"github.com/webond/api/internal/db"
	"github.com/webond/api/internal/money"
	"github.com/webond/api/internal/repo"
)

// PortfolioService computes holdings valuation + unrealized P/L and the
// dashboard aggregates (components.md § PortfolioService/DashboardService;
// FR-21/22, BR-10). MF is valued at latest NAV, Bond at latest clean price,
// Gold at BID (US-022).
type PortfolioService struct {
	store *repo.Store
	clock Clock
}

// NewPortfolioService builds a PortfolioService.
func NewPortfolioService(store *repo.Store, clock Clock) *PortfolioService {
	return &PortfolioService{store: store, clock: clock}
}

// HoldingView is one valued holding row (US-022).
type HoldingView struct {
	AssetType     string `json:"assetType"`
	AssetID       uint64 `json:"assetId"`
	Code          string `json:"code"`
	Name          string `json:"name"`
	Quantity      string `json:"quantity"`
	AvgCost       string `json:"avgCost"`
	MarketPrice   string `json:"marketPrice"`
	MarketValue   string `json:"marketValue"`
	UnrealizedPL  string `json:"unrealizedPl"`
	UnrealizedPct string `json:"unrealizedPct"`
	TotalCost     string `json:"totalCost"`
	NextCoupon    string `json:"nextCouponDate,omitempty"`
	MaturityDate  string `json:"maturityDate,omitempty"`
	AccruedInt    string `json:"accruedInterest,omitempty"`

	marketValue decimal.Decimal
	totalCost   decimal.Decimal
}

// PortfolioSummary is the summary-card block (US-022).
type PortfolioSummary struct {
	MarketValue   string `json:"marketValue"`
	TotalCost     string `json:"totalCost"`
	UnrealizedPL  string `json:"unrealizedPl"`
	RealizedPLYTD string `json:"realizedPlYtd"`
}

// Portfolio is the full portfolio view.
type Portfolio struct {
	Summary  PortfolioSummary `json:"summary"`
	Holdings []HoldingView    `json:"holdings"`
}

// Get returns the user's valued portfolio (US-022, BR-10).
func (s *PortfolioService) Get(ctx context.Context, userID uint64) (Portfolio, error) {
	holdings, err := s.store.Holdings.ListByUser(ctx, userID)
	if err != nil {
		return Portfolio{}, fmt.Errorf("list holdings: %w", err)
	}
	var totalMV, totalCost decimal.Decimal
	views := make([]HoldingView, 0, len(holdings))
	for _, h := range holdings {
		v := s.valueHolding(ctx, h)
		totalMV = totalMV.Add(v.marketValue)
		totalCost = totalCost.Add(v.totalCost)
		views = append(views, v)
	}
	unrealized := money.Quantize(totalMV.Sub(totalCost))
	realizedYTD := s.realizedPLYTD(ctx, userID)
	return Portfolio{
		Summary: PortfolioSummary{
			MarketValue:   money.FormatTHB(totalMV),
			TotalCost:     money.FormatTHB(totalCost),
			UnrealizedPL:  money.FormatTHB(unrealized),
			RealizedPLYTD: money.FormatTHB(realizedYTD),
		},
		Holdings: views,
	}, nil
}

// valueHolding prices a single holding and computes unrealized P/L.
func (s *PortfolioService) valueHolding(ctx context.Context, h db.Holding) HoldingView {
	v := HoldingView{
		AssetType: string(h.AssetType), AssetID: h.AssetID,
		Quantity: money.FormatQuantity(h.Quantity), AvgCost: money.String(h.AvgCost),
		totalCost: h.TotalCost,
	}
	var price decimal.Decimal
	switch h.AssetType {
	case db.HoldingsAssetTypeFund:
		if f, err := s.store.Funds.GetByID(ctx, h.AssetID); err == nil {
			v.Code, v.Name = f.Code, f.Name
		}
		if nav, err := s.store.Funds.GetLatestNav(ctx, h.AssetID); err == nil {
			price = nav.Nav
		}
	case db.HoldingsAssetTypeBond:
		var face decimal.Decimal
		if b, err := s.store.Bonds.GetByID(ctx, h.AssetID); err == nil {
			v.Code, v.Name = b.Code, b.Name
			face = b.FaceValue
			if b.MaturityDate.Valid {
				v.MaturityDate = b.MaturityDate.Time.Format("2006-01-02")
			}
		}
		if pr, err := s.store.Bonds.GetLatestPrice(ctx, h.AssetID); err == nil {
			// clean_price is quoted as % of par; value the holding in THB per
			// unit (US-022: Bond at latest clean price). accrued_interest is
			// already THB per unit and is shown separately.
			price = money.BondUnitPrice(face, pr.CleanPrice)
			v.AccruedInt = money.String(pr.AccruedInterest.Mul(h.Quantity).Round(4))
		}
		v.NextCoupon = s.nextCouponDate(ctx, h.AssetID)
	case db.HoldingsAssetTypeGold:
		if g, err := s.store.Gold.GetByID(ctx, h.AssetID); err == nil {
			v.Code, v.Name = g.Code, g.Name
		}
		if pr, err := s.store.Gold.GetLatestPrice(ctx, h.AssetID); err == nil {
			price = pr.BidPerGram // value gold at BID (US-022)
		}
	}
	mv := money.Quantize(h.Quantity.Mul(price))
	v.marketValue = mv
	v.MarketPrice = money.String(price)
	v.MarketValue = money.FormatTHB(mv)
	upl := money.Quantize(mv.Sub(h.TotalCost))
	v.UnrealizedPL = money.FormatTHB(upl)
	v.TotalCost = money.FormatTHB(h.TotalCost)
	if h.TotalCost.GreaterThan(decimal.Zero) {
		v.UnrealizedPct = upl.Div(h.TotalCost).Mul(decimal.NewFromInt(100)).StringFixed(2)
	} else {
		v.UnrealizedPct = "0.00"
	}
	return v
}

func (s *PortfolioService) nextCouponDate(ctx context.Context, bondID uint64) string {
	coupons, err := s.store.Bonds.ListUnpaidCouponsDue(ctx, s.clock.Now().AddDate(10, 0, 0))
	if err != nil {
		return ""
	}
	var next time.Time
	for _, c := range coupons {
		if c.BondID != bondID {
			continue
		}
		if c.CouponDate.After(s.clock.Now()) && (next.IsZero() || c.CouponDate.Before(next)) {
			next = c.CouponDate
		}
	}
	if next.IsZero() {
		return ""
	}
	return next.Format("2006-01-02")
}

// realizedPLYTD sums realized P/L from this year's transactions.
func (s *PortfolioService) realizedPLYTD(ctx context.Context, userID uint64) decimal.Decimal {
	rows, err := s.store.Transactions.ListByUser(ctx, db.ListTransactionsByUserParams{
		UserID: userID, Limit: 1000, Offset: 0,
	})
	if err != nil {
		return decimal.Zero
	}
	yearStart := time.Date(s.clock.Now().Year(), 1, 1, 0, 0, 0, 0, Bangkok)
	total := decimal.Zero
	for _, r := range rows {
		if r.RealizedPl.Valid && r.OccurredAt.After(yearStart) {
			total = total.Add(r.RealizedPl.Decimal)
		}
	}
	return money.Quantize(total)
}
