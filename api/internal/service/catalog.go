package service

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/shopspring/decimal"

	"github.com/webond/api/internal/db"
	"github.com/webond/api/internal/domain"
	"github.com/webond/api/internal/money"
	"github.com/webond/api/internal/repo"
)

// CatalogService implements the market catalog: list/search/filter assets,
// multi-period returns + performance sort (Add-on), type-specific detail, and
// price/NAV history for charts (components.md § CatalogService; FR-05/06,
// US-011/012).
type CatalogService struct {
	store *repo.Store
	clock Clock
}

// NewCatalogService builds a CatalogService.
func NewCatalogService(store *repo.Store, clock Clock) *CatalogService {
	return &CatalogService{store: store, clock: clock}
}

// Staleness thresholds (EX-10, US-012): gold price older than 5 min and NAV
// older than 1 business day are treated as "price unavailable".
const (
	goldStaleAfter = 5 * time.Minute
	navStaleHours  = 36 * time.Hour // ~1 business day allowance
)

// AssetListItem is one row of the market list (US-011).
type AssetListItem struct {
	Type         string          `json:"type"`
	ID           uint64          `json:"id"`
	Code         string          `json:"code"`
	Name         string          `json:"name"`
	Category     string          `json:"category,omitempty"`
	RiskLevel    int16           `json:"riskLevel,omitempty"`
	LatestPrice  string          `json:"latestPrice,omitempty"`  // NAV / clean price / gold ASK
	BidPrice     string          `json:"bidPrice,omitempty"`     // gold BID
	AskPrice     string          `json:"askPrice,omitempty"`     // gold ASK
	Ytm          string          `json:"ytm,omitempty"`          // bond
	MaturityDate string          `json:"maturityDate,omitempty"` // bond
	PeriodReturn string          `json:"periodReturn,omitempty"` // % over selected period
	periodReturn decimal.Decimal // sort key (unexported)
	PriceAsOf    string          `json:"priceAsOf,omitempty"`
}

// ListParams are the market list filters (US-011 Add-on).
type ListParams struct {
	Type      string // fund|bond|gold ("" = all)
	Query     string
	RiskLevel int
	Category  string
	Period    string // 3M|6M|1Y|3Y|5Y
	Sort      string // performance|name
	Order     string // asc|desc
	Page      int
	PageSize  int
}

// List returns market assets with latest price + period return (US-011).
func (s *CatalogService) List(ctx context.Context, p ListParams) ([]AssetListItem, error) {
	var items []AssetListItem
	types := []string{p.Type}
	if p.Type == "" {
		types = []string{"fund", "bond", "gold"}
	}
	for _, typ := range types {
		switch typ {
		case "fund":
			fs, err := s.listFunds(ctx, p)
			if err != nil {
				return nil, err
			}
			items = append(items, fs...)
		case "bond":
			bs, err := s.listBonds(ctx, p)
			if err != nil {
				return nil, err
			}
			items = append(items, bs...)
		case "gold":
			gs, err := s.listGold(ctx, p)
			if err != nil {
				return nil, err
			}
			items = append(items, gs...)
		default:
			return nil, domain.NewError(domain.CodeValidationError, "Unknown asset type")
		}
	}

	items = filterAndSearch(items, p)
	sortAssets(items, p)
	return items, nil
}

func (s *CatalogService) listFunds(ctx context.Context, p ListParams) ([]AssetListItem, error) {
	rows, err := s.store.Funds.List(ctx, db.ListFundsParams{Limit: 500, Offset: 0})
	if err != nil {
		return nil, fmt.Errorf("list funds: %w", err)
	}
	out := make([]AssetListItem, 0, len(rows))
	for _, f := range rows {
		item := AssetListItem{
			Type: "fund", ID: f.ID, Code: f.Code, Name: f.Name,
			Category: f.Category.String,
		}
		if f.RiskLevel.Valid {
			item.RiskLevel = f.RiskLevel.Int16
		}
		latest, lerr := s.store.Funds.GetLatestNav(ctx, f.ID)
		if lerr == nil {
			item.LatestPrice = money.String(latest.Nav)
			item.PriceAsOf = latest.NavDate.Format("2006-01-02")
			item.periodReturn, item.PeriodReturn = s.fundPeriodReturn(ctx, f.ID, latest.Nav, p.Period)
		}
		out = append(out, item)
	}
	return out, nil
}

func (s *CatalogService) listBonds(ctx context.Context, p ListParams) ([]AssetListItem, error) {
	rows, err := s.store.Bonds.List(ctx, db.ListBondsParams{Limit: 500, Offset: 0})
	if err != nil {
		return nil, fmt.Errorf("list bonds: %w", err)
	}
	out := make([]AssetListItem, 0, len(rows))
	for _, b := range rows {
		item := AssetListItem{
			Type: "bond", ID: b.ID, Code: b.Code, Name: b.Name,
		}
		if b.RiskLevel.Valid {
			item.RiskLevel = b.RiskLevel.Int16
		}
		if b.Ytm.Valid {
			item.Ytm = money.String(b.Ytm.Decimal)
		}
		if b.MaturityDate.Valid {
			item.MaturityDate = b.MaturityDate.Time.Format("2006-01-02")
		}
		latest, lerr := s.store.Bonds.GetLatestPrice(ctx, b.ID)
		if lerr == nil {
			// THB per unit (face x clean% / 100), the unit used by the portfolio
			// and orders; the period return is the same either way.
			item.LatestPrice = money.String(money.BondUnitPrice(b.FaceValue, latest.CleanPrice))
			item.PriceAsOf = latest.PriceDate.Format("2006-01-02")
			item.periodReturn, item.PeriodReturn = s.bondPeriodReturn(ctx, b.ID, latest.CleanPrice, p.Period)
		}
		out = append(out, item)
	}
	return out, nil
}

func (s *CatalogService) listGold(ctx context.Context, p ListParams) ([]AssetListItem, error) {
	rows, err := s.store.Gold.List(ctx, db.ListGoldProductsParams{Limit: 500, Offset: 0})
	if err != nil {
		return nil, fmt.Errorf("list gold: %w", err)
	}
	out := make([]AssetListItem, 0, len(rows))
	for _, g := range rows {
		item := AssetListItem{Type: "gold", ID: g.ID, Code: g.Code, Name: g.Name}
		latest, lerr := s.store.Gold.GetLatestPrice(ctx, g.ID)
		if lerr == nil {
			item.BidPrice = money.String(latest.BidPerGram)
			item.AskPrice = money.String(latest.AskPerGram)
			item.LatestPrice = money.String(latest.AskPerGram)
			item.PriceAsOf = latest.PricedAt.Format("2006-01-02T15:04:05-07:00")
			item.periodReturn, item.PeriodReturn = s.goldPeriodReturn(ctx, g.ID, latest.AskPerGram, p.Period)
		}
		out = append(out, item)
	}
	return out, nil
}

// periodStart maps a period code to a start date relative to now.
func (s *CatalogService) periodStart(period string) time.Time {
	now := s.clock.Now()
	switch strings.ToUpper(period) {
	case "3M":
		return now.AddDate(0, -3, 0)
	case "6M":
		return now.AddDate(0, -6, 0)
	case "1Y":
		return now.AddDate(-1, 0, 0)
	case "3Y":
		return now.AddDate(-3, 0, 0)
	case "5Y":
		return now.AddDate(-5, 0, 0)
	case "1M":
		return now.AddDate(0, -1, 0)
	case "ALL", "":
		return time.Date(1970, 1, 1, 0, 0, 0, 0, Bangkok)
	default:
		return now.AddDate(-1, 0, 0)
	}
}

// pctReturn computes (latest−base)/base * 100, quantized to 4dp. Zero base → 0.
func pctReturn(base, latest decimal.Decimal) (decimal.Decimal, string) {
	if base.LessThanOrEqual(decimal.Zero) {
		return decimal.Zero, ""
	}
	r := money.Quantize(latest.Sub(base).Div(base).Mul(decimal.NewFromInt(100)))
	return r, r.StringFixed(2)
}

func (s *CatalogService) fundPeriodReturn(ctx context.Context, fundID uint64, latest decimal.Decimal, period string) (decimal.Decimal, string) {
	navs, err := s.store.Funds.ListNavs(ctx, db.ListFundNavsParams{FundID: fundID, NavDate: s.periodStart(period)})
	if err != nil || len(navs) == 0 {
		return decimal.Zero, ""
	}
	return pctReturn(navs[0].Nav, latest)
}

func (s *CatalogService) bondPeriodReturn(ctx context.Context, bondID uint64, latest decimal.Decimal, period string) (decimal.Decimal, string) {
	prices, err := s.store.Bonds.ListPrices(ctx, db.ListBondPricesParams{BondID: bondID, PriceDate: s.periodStart(period)})
	if err != nil || len(prices) == 0 {
		return decimal.Zero, ""
	}
	return pctReturn(prices[0].CleanPrice, latest)
}

func (s *CatalogService) goldPeriodReturn(ctx context.Context, goldID uint64, latest decimal.Decimal, period string) (decimal.Decimal, string) {
	prices, err := s.store.Gold.ListPrices(ctx, db.ListGoldPricesParams{GoldProductID: goldID, PricedAt: s.periodStart(period)})
	if err != nil || len(prices) == 0 {
		return decimal.Zero, ""
	}
	return pctReturn(prices[0].AskPerGram, latest)
}

// filterAndSearch applies q/riskLevel/category filters in memory.
func filterAndSearch(items []AssetListItem, p ListParams) []AssetListItem {
	q := strings.ToLower(strings.TrimSpace(p.Query))
	out := items[:0]
	for _, it := range items {
		if q != "" && !strings.Contains(strings.ToLower(it.Code), q) && !strings.Contains(strings.ToLower(it.Name), q) {
			continue
		}
		if p.RiskLevel > 0 && int(it.RiskLevel) != p.RiskLevel {
			continue
		}
		if p.Category != "" && !strings.EqualFold(it.Category, p.Category) {
			continue
		}
		out = append(out, it)
	}
	return out
}

// sortAssets orders by performance (period return) or name (US-011).
func sortAssets(items []AssetListItem, p ListParams) {
	desc := strings.EqualFold(p.Order, "desc")
	switch strings.ToLower(p.Sort) {
	case "performance":
		sort.SliceStable(items, func(i, j int) bool {
			c := items[i].periodReturn.Cmp(items[j].periodReturn)
			if desc {
				return c > 0
			}
			return c < 0
		})
	default: // name
		sort.SliceStable(items, func(i, j int) bool {
			if desc {
				return items[i].Code > items[j].Code
			}
			return items[i].Code < items[j].Code
		})
	}
}

// AssetDetail is the type-specific detail payload (US-012).
type AssetDetail struct {
	Type        string         `json:"type"`
	ID          uint64         `json:"id"`
	Code        string         `json:"code"`
	Name        string         `json:"name"`
	Fields      map[string]any `json:"fields"`
	PriceAsOf   string         `json:"priceAsOf,omitempty"`
	Unavailable bool           `json:"priceUnavailable"`
}

// Detail returns type-specific asset detail (US-012). Staleness is flagged per
// EX-10 (gold 5 min, NAV 1 business day).
func (s *CatalogService) Detail(ctx context.Context, assetType string, id uint64) (AssetDetail, error) {
	switch assetType {
	case "fund":
		f, err := s.store.Funds.GetByID(ctx, id)
		if err != nil {
			return AssetDetail{}, assetNotFound(err)
		}
		d := AssetDetail{Type: "fund", ID: f.ID, Code: f.Code, Name: f.Name, Fields: map[string]any{
			"category": f.Category.String, "riskLevel": f.RiskLevel.Int16,
			"minPurchase": money.String(f.MinPurchase), "settlementDays": f.SettlementDays,
			"cutoffTime": f.CutoffTime, "fee": "0.0000",
			"factSheetUrl": f.FactSheetUrl.String,
		}}
		if nav, nerr := s.store.Funds.GetLatestNav(ctx, f.ID); nerr == nil {
			d.Fields["nav"] = money.String(nav.Nav)
			d.PriceAsOf = nav.NavDate.Format("2006-01-02")
			d.Unavailable = s.clock.Now().Sub(nav.NavDate) > navStaleHours
		} else {
			d.Unavailable = true
		}
		return d, nil
	case "bond":
		b, err := s.store.Bonds.GetByID(ctx, id)
		if err != nil {
			return AssetDetail{}, assetNotFound(err)
		}
		d := AssetDetail{Type: "bond", ID: b.ID, Code: b.Code, Name: b.Name, Fields: map[string]any{
			"issuer": b.Issuer.String, "couponRate": money.String(b.CouponRate),
			"couponFrequency": b.CouponFrequency, "faceValue": money.String(b.FaceValue),
			"ytm": decimalOrEmpty(b.Ytm), "creditRating": b.CreditRating.String,
			"minLot": money.String(b.MinLot), "riskLevel": b.RiskLevel.Int16,
		}}
		if b.MaturityDate.Valid {
			d.Fields["maturityDate"] = b.MaturityDate.Time.Format("2006-01-02")
		}
		if b.OfferingStart.Valid {
			d.Fields["offeringStart"] = b.OfferingStart.Time.Format("2006-01-02")
		}
		if b.OfferingEnd.Valid {
			d.Fields["offeringEnd"] = b.OfferingEnd.Time.Format("2006-01-02")
		}
		if pr, perr := s.store.Bonds.GetLatestPrice(ctx, b.ID); perr == nil {
			d.Fields["cleanPrice"] = money.String(pr.CleanPrice) // % of par
			d.Fields["price"] = money.String(money.BondUnitPrice(b.FaceValue, pr.CleanPrice))
			d.Fields["accruedInterest"] = money.String(pr.AccruedInterest)
			d.PriceAsOf = pr.PriceDate.Format("2006-01-02")
		}
		return d, nil
	case "gold":
		g, err := s.store.Gold.GetByID(ctx, id)
		if err != nil {
			return AssetDetail{}, assetNotFound(err)
		}
		d := AssetDetail{Type: "gold", ID: g.ID, Code: g.Code, Name: g.Name, Fields: map[string]any{
			"purity": g.Purity.String,
		}}
		if pr, perr := s.store.Gold.GetLatestPrice(ctx, g.ID); perr == nil {
			d.Fields["bid"] = money.String(pr.BidPerGram)
			d.Fields["ask"] = money.String(pr.AskPerGram)
			d.PriceAsOf = pr.PricedAt.Format("2006-01-02T15:04:05-07:00")
			d.Unavailable = s.clock.Now().Sub(pr.PricedAt) > goldStaleAfter
		} else {
			d.Unavailable = true
		}
		return d, nil
	default:
		return AssetDetail{}, domain.NewError(domain.CodeValidationError, "Unknown asset type")
	}
}

// PricePoint is one history series point for charts (US-012).
type PricePoint struct {
	Date  string `json:"date"`
	Value string `json:"value"`
	Bid   string `json:"bid,omitempty"`
	Ask   string `json:"ask,omitempty"`
}

// PriceHistory returns the price/NAV series for an asset over a period (US-012).
func (s *CatalogService) PriceHistory(ctx context.Context, assetType string, id uint64, period string) ([]PricePoint, error) {
	start := s.periodStart(period)
	switch assetType {
	case "fund":
		navs, err := s.store.Funds.ListNavs(ctx, db.ListFundNavsParams{FundID: id, NavDate: start})
		if err != nil {
			return nil, fmt.Errorf("list navs: %w", err)
		}
		out := make([]PricePoint, 0, len(navs))
		for _, n := range navs {
			out = append(out, PricePoint{Date: n.NavDate.Format("2006-01-02"), Value: money.String(n.Nav)})
		}
		return out, nil
	case "bond":
		bond, err := s.store.Bonds.GetByID(ctx, id)
		if err != nil {
			return nil, assetNotFound(err)
		}
		prices, err := s.store.Bonds.ListPrices(ctx, db.ListBondPricesParams{BondID: id, PriceDate: start})
		if err != nil {
			return nil, fmt.Errorf("list bond prices: %w", err)
		}
		out := make([]PricePoint, 0, len(prices))
		for _, pr := range prices {
			out = append(out, PricePoint{
				Date:  pr.PriceDate.Format("2006-01-02"),
				Value: money.String(money.BondUnitPrice(bond.FaceValue, pr.CleanPrice)),
			})
		}
		return out, nil
	case "gold":
		prices, err := s.store.Gold.ListPrices(ctx, db.ListGoldPricesParams{GoldProductID: id, PricedAt: start})
		if err != nil {
			return nil, fmt.Errorf("list gold prices: %w", err)
		}
		out := make([]PricePoint, 0, len(prices))
		for _, pr := range prices {
			out = append(out, PricePoint{
				Date:  pr.PricedAt.Format("2006-01-02T15:04:05-07:00"),
				Value: money.String(pr.AskPerGram), Bid: money.String(pr.BidPerGram), Ask: money.String(pr.AskPerGram),
			})
		}
		return out, nil
	default:
		return nil, domain.NewError(domain.CodeValidationError, "Unknown asset type")
	}
}

func assetNotFound(err error) error {
	if isNoRows(err) {
		return domain.NewError(domain.CodeTxnNotFound, "Asset not found")
	}
	return fmt.Errorf("get asset: %w", err)
}

func decimalOrEmpty(d decimal.NullDecimal) string {
	if !d.Valid {
		return ""
	}
	return money.String(d.Decimal)
}
