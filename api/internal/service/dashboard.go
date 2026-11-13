package service

import (
	"context"
	"sort"
	"time"

	"github.com/shopspring/decimal"

	"github.com/webond/api/internal/db"
	"github.com/webond/api/internal/money"
	"github.com/webond/api/internal/repo"
)

// DashboardService builds the consolidated insight dashboard (components.md §
// DashboardService; FR-22, US-023). It reuses PortfolioService for valuation.
type DashboardService struct {
	store     *repo.Store
	portfolio *PortfolioService
	wallet    *WalletService
	clock     Clock
}

// NewDashboardService builds a DashboardService.
func NewDashboardService(store *repo.Store, portfolio *PortfolioService, wallet *WalletService, clock Clock) *DashboardService {
	return &DashboardService{store: store, portfolio: portfolio, wallet: wallet, clock: clock}
}

// AllocationSlice is one allocation-pie slice (US-023).
type AllocationSlice struct {
	AssetType string `json:"assetType"`
	Value     string `json:"value"`
	Pct       string `json:"pct"`
}

// TrendPoint is one value-trend series point (US-023 Add-on).
type TrendPoint struct {
	Date  string `json:"date"`
	Value string `json:"value"`
}

// Mover is a top gainer/loser row (US-023).
type Mover struct {
	AssetType string `json:"assetType"`
	AssetID   uint64 `json:"assetId"`
	Code      string `json:"code"`
	Name      string `json:"name"`
	Pct       string `json:"pct"`
}

// UpcomingEvent is a coupon/maturity/DCA-run reminder (US-023).
type UpcomingEvent struct {
	Type   string `json:"type"` // coupon | maturity | dca_run
	Date   string `json:"date"`
	Ref    string `json:"ref"`              // asset code
	Amount string `json:"amount,omitempty"` // expected cash amount (THB) when known
}

// Dashboard is the full dashboard payload.
type Dashboard struct {
	TotalValue     string            `json:"totalValue"`
	TotalPL        string            `json:"totalPl"`
	CashAvailable  string            `json:"cashAvailable"`
	Allocation     []AllocationSlice `json:"allocation"`
	Trend          []TrendPoint      `json:"trend"`
	TopMovers      []Mover           `json:"topMovers"`
	DCASummary     DCASummary        `json:"dcaSummary"`
	UpcomingEvents []UpcomingEvent   `json:"upcomingEvents"`
	LastUpdated    string            `json:"lastUpdated"`
}

// DCASummary aggregates the user's DCA activity (US-023).
type DCASummary struct {
	ActivePlans   int    `json:"activePlans"`
	TotalInvested string `json:"totalInvested"`
	// MonthlyCommitment is what the active plans invest per month: monthly plans
	// at face value, weekly plans x 52 / 12.
	MonthlyCommitment string `json:"monthlyCommitment"`
}

// Get builds the dashboard for a user over the requested trend range (US-023).
func (s *DashboardService) Get(ctx context.Context, userID uint64, trend string) (Dashboard, error) {
	pf, err := s.portfolio.Get(ctx, userID)
	if err != nil {
		return Dashboard{}, err
	}
	bal, _ := s.wallet.Get(ctx, userID)

	// Totals.
	totalValue := decimalFromTHB(pf.Summary.MarketValue)
	totalPL := decimalFromTHB(pf.Summary.UnrealizedPL)

	d := Dashboard{
		TotalValue:    pf.Summary.MarketValue,
		TotalPL:       money.FormatTHB(totalPL),
		CashAvailable: money.FormatTHB(bal.Available),
		LastUpdated:   s.clock.Now().Format("2006-01-02T15:04:05-07:00"),
	}

	// Allocation by asset type.
	d.Allocation = allocation(pf.Holdings, totalValue)
	// Top movers (by unrealized %).
	d.TopMovers = topMovers(pf.Holdings)
	// DCA summary.
	d.DCASummary = s.dcaSummary(ctx, userID)
	// Upcoming events (coupons, maturities, DCA runs).
	d.UpcomingEvents = s.upcomingEvents(ctx, userID)
	// Value trend series.
	d.Trend = s.valueTrend(ctx, userID, trend)
	return d, nil
}

func allocation(holdings []HoldingView, total decimal.Decimal) []AllocationSlice {
	byType := map[string]decimal.Decimal{}
	for _, h := range holdings {
		byType[h.AssetType] = byType[h.AssetType].Add(h.marketValue)
	}
	out := make([]AllocationSlice, 0, len(byType))
	for _, typ := range []string{"fund", "bond", "gold"} {
		v, ok := byType[typ]
		if !ok {
			continue
		}
		pct := "0.00"
		if total.GreaterThan(decimal.Zero) {
			pct = v.Div(total).Mul(decimal.NewFromInt(100)).StringFixed(2)
		}
		out = append(out, AllocationSlice{AssetType: typ, Value: money.FormatTHB(v), Pct: pct})
	}
	return out
}

func topMovers(holdings []HoldingView) []Mover {
	movers := make([]Mover, 0, len(holdings))
	for _, h := range holdings {
		movers = append(movers, Mover{AssetType: h.AssetType, AssetID: h.AssetID, Code: h.Code, Name: h.Name, Pct: h.UnrealizedPct})
	}
	sort.SliceStable(movers, func(i, j int) bool {
		return decimalFromString(movers[i].Pct).GreaterThan(decimalFromString(movers[j].Pct))
	})
	if len(movers) > 5 {
		movers = movers[:5]
	}
	return movers
}

func (s *DashboardService) dcaSummary(ctx context.Context, userID uint64) DCASummary {
	plans, err := s.store.Dca.ListPlansByUser(ctx, userID)
	if err != nil {
		return DCASummary{}
	}
	active := 0
	invested := decimal.Zero
	monthly := decimal.Zero
	for _, p := range plans {
		if p.Status == db.DcaPlansStatusActive {
			active++
			monthly = monthly.Add(monthlyAmount(p.Frequency, p.AmountPerRound))
		}
		invested = invested.Add(p.TotalInvested)
	}
	return DCASummary{
		ActivePlans:       active,
		TotalInvested:     money.FormatTHB(money.Quantize(invested)),
		MonthlyCommitment: money.FormatTHB(money.Quantize(monthly)),
	}
}

// monthlyAmount converts a plan's per-round amount into a per-month amount.
func monthlyAmount(freq db.DcaPlansFrequency, perRound decimal.Decimal) decimal.Decimal {
	if freq == db.DcaPlansFrequencyWeekly {
		return perRound.Mul(decimal.NewFromInt(52)).DivRound(decimal.NewFromInt(12), money.Scale+8)
	}
	return perRound
}

func (s *DashboardService) upcomingEvents(ctx context.Context, userID uint64) []UpcomingEvent {
	events := []UpcomingEvent{}
	now := s.clock.Now()

	// Coupons & maturities for held bonds within the next 90 days.
	holdings, _ := s.store.Holdings.ListByUser(ctx, userID)
	type heldBond struct {
		code string
		qty  decimal.Decimal
	}
	heldBonds := map[uint64]heldBond{}
	for _, h := range holdings {
		if h.AssetType == db.HoldingsAssetTypeBond && h.Quantity.GreaterThan(decimal.Zero) {
			b, err := s.store.Bonds.GetByID(ctx, h.AssetID)
			if err != nil {
				continue
			}
			heldBonds[h.AssetID] = heldBond{code: b.Code, qty: h.Quantity}
			if b.MaturityDate.Valid && b.MaturityDate.Time.After(now) && b.MaturityDate.Time.Before(now.AddDate(0, 3, 0)) {
				events = append(events, UpcomingEvent{
					Type: "maturity", Date: b.MaturityDate.Time.Format("2006-01-02"), Ref: b.Code,
					Amount: money.FormatTHB(money.BondFaceValue(h.Quantity, b.FaceValue)),
				})
			}
		}
	}
	coupons, _ := s.store.Bonds.ListUnpaidCouponsDue(ctx, now.AddDate(0, 3, 0))
	for _, c := range coupons {
		if hb, ok := heldBonds[c.BondID]; ok && c.CouponDate.After(now) {
			events = append(events, UpcomingEvent{
				Type: "coupon", Date: c.CouponDate.Format("2006-01-02"), Ref: hb.code,
				Amount: money.FormatTHB(money.Quantize(c.AmountPerUnit.Mul(hb.qty))),
			})
		}
	}

	// Next DCA runs.
	plans, _ := s.store.Dca.ListPlansByUser(ctx, userID)
	for _, p := range plans {
		if p.Status == db.DcaPlansStatusActive && p.NextRunDate.Valid {
			events = append(events, UpcomingEvent{
				Type: "dca_run", Date: p.NextRunDate.Time.Format("2006-01-02"),
				Ref: s.assetCode(ctx, string(p.AssetType), p.AssetID), Amount: money.FormatTHB(p.AmountPerRound),
			})
		}
	}
	sort.SliceStable(events, func(i, j int) bool { return events[i].Date < events[j].Date })
	return events
}

// valueTrend builds a daily portfolio-value series over the range using the
// fund NAV / bond price / gold price history for current holdings (US-023 Add-on).
// Phase-1 approximation: current holding quantity × historical price.
func (s *DashboardService) valueTrend(ctx context.Context, userID uint64, trend string) []TrendPoint {
	start := s.trendStart(trend)
	holdings, _ := s.store.Holdings.ListByUser(ctx, userID)
	// value per day keyed by date string.
	byDay := map[string]decimal.Decimal{}
	for _, h := range holdings {
		switch h.AssetType {
		case db.HoldingsAssetTypeFund:
			navs, _ := s.store.Funds.ListNavs(ctx, db.ListFundNavsParams{FundID: h.AssetID, NavDate: start})
			for _, n := range navs {
				k := n.NavDate.Format("2006-01-02")
				byDay[k] = byDay[k].Add(h.Quantity.Mul(n.Nav))
			}
		case db.HoldingsAssetTypeBond:
			bond, berr := s.store.Bonds.GetByID(ctx, h.AssetID)
			if berr != nil {
				continue // cannot value a bond without its face value
			}
			prices, _ := s.store.Bonds.ListPrices(ctx, db.ListBondPricesParams{BondID: h.AssetID, PriceDate: start})
			for _, p := range prices {
				k := p.PriceDate.Format("2006-01-02")
				// clean_price is % of par -> THB per unit via the face value.
				byDay[k] = byDay[k].Add(h.Quantity.Mul(money.BondUnitPrice(bond.FaceValue, p.CleanPrice)))
			}
		case db.HoldingsAssetTypeGold:
			prices, _ := s.store.Gold.ListPrices(ctx, db.ListGoldPricesParams{GoldProductID: h.AssetID, PricedAt: start})
			// Gold ticks many times a day (live price generator): value each
			// day with its LAST price only, never the sum of all ticks.
			lastAt := map[string]time.Time{}
			lastBid := map[string]decimal.Decimal{}
			for _, p := range prices {
				k := p.PricedAt.In(Bangkok).Format("2006-01-02")
				if t, seen := lastAt[k]; !seen || p.PricedAt.After(t) {
					lastAt[k], lastBid[k] = p.PricedAt, p.BidPerGram
				}
			}
			for k, bid := range lastBid {
				byDay[k] = byDay[k].Add(h.Quantity.Mul(bid))
			}
		}
	}
	dates := make([]string, 0, len(byDay))
	for k := range byDay {
		dates = append(dates, k)
	}
	sort.Strings(dates)
	out := make([]TrendPoint, 0, len(dates))
	for _, d := range dates {
		out = append(out, TrendPoint{Date: d, Value: money.FormatTHB(money.Quantize(byDay[d]))})
	}
	return out
}

func (s *DashboardService) trendStart(trend string) time.Time {
	now := s.clock.Now()
	switch trend {
	case "1M":
		return now.AddDate(0, -1, 0)
	case "6M":
		return now.AddDate(0, -6, 0)
	case "1Y":
		return now.AddDate(-1, 0, 0)
	case "3Y":
		return now.AddDate(-3, 0, 0)
	case "5Y":
		return now.AddDate(-5, 0, 0)
	case "all", "All", "":
		return time.Date(1970, 1, 1, 0, 0, 0, 0, Bangkok)
	default:
		return now.AddDate(0, -1, 0)
	}
}

func decimalFromTHB(s string) decimal.Decimal { return decimalFromString(s) }

func decimalFromString(s string) decimal.Decimal {
	d, err := decimal.NewFromString(s)
	if err != nil {
		return decimal.Zero
	}
	return d
}

// assetCode returns the short code of a fund or gold product ("" if unknown).
func (s *DashboardService) assetCode(ctx context.Context, assetType string, id uint64) string {
	switch assetType {
	case "fund":
		if f, err := s.store.Funds.GetByID(ctx, id); err == nil {
			return f.Code
		}
	case "gold":
		if g, err := s.store.Gold.GetByID(ctx, id); err == nil {
			return g.Code
		}
	}
	return ""
}
