package service

import (
	"context"
	"log/slog"
	"math/rand"
	"sync"
	"time"

	"github.com/shopspring/decimal"

	"github.com/webond/api/internal/db"
	"github.com/webond/api/internal/money"
	"github.com/webond/api/internal/repo"
)

// PriceGeneratorService simulates market movement by periodically appending new
// gold prices, fund NAVs and bond prices to the DB (components.md § MockExternal
// + PriceGenerator; US-030 Add-on, integration.md). The on/off toggle + config
// are persisted in price_generator_state (singleton) so they survive restarts
// (D3-7). The frontend polls latest prices (no WebSocket in phase 1).
type PriceGeneratorService struct {
	store  *repo.Store
	logger *slog.Logger
	clock  Clock

	mu     sync.Mutex
	cancel context.CancelFunc
	rng    *rand.Rand
}

// NewPriceGeneratorService builds a PriceGeneratorService.
func NewPriceGeneratorService(store *repo.Store, logger *slog.Logger, clock Clock) *PriceGeneratorService {
	return &PriceGeneratorService{
		store: store, logger: logger, clock: clock,
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// State is the generator state view (US-030).
type State struct {
	Enabled         bool   `json:"enabled"`
	IntervalSeconds int32  `json:"intervalSeconds"`
	VolatilityPct   string `json:"volatilityPct"`
	LastRunAt       string `json:"lastRunAt,omitempty"`
}

// GetState returns the persisted generator state (US-030).
func (s *PriceGeneratorService) GetState(ctx context.Context) (State, error) {
	st, err := s.store.PriceGen.Get(ctx)
	if err != nil {
		return State{}, err
	}
	return stateView(st), nil
}

// SetState persists the toggle/config and starts or stops the background ticker
// accordingly (US-030, D3-7).
func (s *PriceGeneratorService) SetState(ctx context.Context, enabled bool, intervalSeconds int, volatilityPct decimal.Decimal) (State, error) {
	if intervalSeconds <= 0 {
		intervalSeconds = 60
	}
	if volatilityPct.LessThanOrEqual(decimal.Zero) {
		volatilityPct = decimal.RequireFromString("0.5")
	}
	cur, err := s.store.PriceGen.Get(ctx)
	if err != nil {
		return State{}, err
	}
	if err := s.store.PriceGen.Update(ctx, db.UpdatePriceGeneratorStateParams{
		Enabled: enabled, IntervalSeconds: int32(intervalSeconds),
		VolatilityPct: volatilityPct, LastRunAt: cur.LastRunAt,
	}); err != nil {
		return State{}, err
	}
	if enabled {
		s.start(time.Duration(intervalSeconds)*time.Second, volatilityPct)
	} else {
		s.Stop()
	}
	st, _ := s.store.PriceGen.Get(ctx)
	return stateView(st), nil
}

// StartIfEnabled launches the ticker at boot if the persisted state is enabled
// (survives restarts — D3-7). Called from main after wiring.
func (s *PriceGeneratorService) StartIfEnabled(ctx context.Context) {
	st, err := s.store.PriceGen.Get(ctx)
	if err != nil {
		s.logger.Warn("price generator: cannot read state at boot", "error", err)
		return
	}
	if st.Enabled {
		s.start(time.Duration(st.IntervalSeconds)*time.Second, st.VolatilityPct)
	}
}

// start launches (or restarts) the background ticker goroutine.
func (s *PriceGeneratorService) start(interval time.Duration, volatility decimal.Decimal) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cancel != nil {
		s.cancel() // stop any existing ticker before restarting
	}
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	go s.loop(ctx, interval, volatility)
	s.logger.Info("price generator started", "intervalSeconds", interval.Seconds())
}

// Stop halts the background ticker if running.
func (s *PriceGeneratorService) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cancel != nil {
		s.cancel()
		s.cancel = nil
		s.logger.Info("price generator stopped")
	}
}

// loop ticks on the interval, generating a new price set each tick. A failing
// tick is logged and skipped; it never crashes the server (integration.md).
func (s *PriceGeneratorService) loop(ctx context.Context, interval time.Duration, volatility decimal.Decimal) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.Tick(ctx, volatility); err != nil {
				s.logger.Warn("price generator tick failed", "error", err)
			}
		}
	}
}

// Tick generates one simulated price set: appends a new gold price (bounded
// random walk on bid/ask), a new fund NAV and a new bond price for each active
// asset (US-030). It is also exposed for the manual publish/tick endpoints.
func (s *PriceGeneratorService) Tick(ctx context.Context, volatility decimal.Decimal) error {
	now := s.clock.Now()
	if err := s.TickGold(ctx, volatility); err != nil {
		return err
	}
	if err := s.PublishNav(ctx, volatility); err != nil {
		return err
	}
	if err := s.tickBonds(ctx, volatility); err != nil {
		return err
	}
	// Record last-run (best effort).
	if cur, err := s.store.PriceGen.Get(ctx); err == nil {
		_ = s.store.PriceGen.Update(ctx, db.UpdatePriceGeneratorStateParams{
			Enabled: cur.Enabled, IntervalSeconds: cur.IntervalSeconds,
			VolatilityPct: cur.VolatilityPct, LastRunAt: nullTime(now),
		})
	}
	return nil
}

// TickGold appends a new bounded-random-walk BID/ASK for each gold product.
func (s *PriceGeneratorService) TickGold(ctx context.Context, volatility decimal.Decimal) error {
	products, err := s.store.Gold.List(ctx, db.ListGoldProductsParams{Limit: 100, Offset: 0})
	if err != nil {
		return err
	}
	now := s.clock.Now()
	for _, g := range products {
		last, lerr := s.store.Gold.GetLatestPrice(ctx, g.ID)
		bid := decimal.RequireFromString("41000")
		ask := decimal.RequireFromString("41250")
		if lerr == nil {
			bid = s.walk(last.BidPerGram, volatility)
			ask = s.walk(last.AskPerGram, volatility)
			if ask.LessThan(bid) {
				ask = bid.Add(decimal.RequireFromString("50"))
			}
		}
		if _, err := s.store.Gold.CreatePrice(ctx, db.CreateGoldPriceParams{
			GoldProductID: g.ID, PricedAt: now, BidPerGram: bid.Round(4), AskPerGram: ask.Round(4),
		}); err != nil {
			return err
		}
	}
	return nil
}

// PublishNav appends a new NAV point for each fund (manual publish-nav endpoint).
func (s *PriceGeneratorService) PublishNav(ctx context.Context, volatility decimal.Decimal) error {
	funds, err := s.store.Funds.List(ctx, db.ListFundsParams{Limit: 500, Offset: 0})
	if err != nil {
		return err
	}
	today := s.today()
	for _, f := range funds {
		nav := decimal.RequireFromString("10")
		if last, lerr := s.store.Funds.GetLatestNav(ctx, f.ID); lerr == nil {
			nav = s.walk(last.Nav, volatility)
		}
		if err := s.store.Funds.UpsertNav(ctx, db.UpsertFundNavParams{
			FundID: f.ID, NavDate: today, Nav: nav.Round(4),
		}); err != nil {
			return err
		}
	}
	return nil
}

func (s *PriceGeneratorService) tickBonds(ctx context.Context, volatility decimal.Decimal) error {
	bonds, err := s.store.Bonds.List(ctx, db.ListBondsParams{Limit: 500, Offset: 0})
	if err != nil {
		return err
	}
	today := s.today()
	for _, b := range bonds {
		clean := money.BondParPercent // quoted as % of par; start at par (100), not face value
		accrued := decimal.Zero
		if last, lerr := s.store.Bonds.GetLatestPrice(ctx, b.ID); lerr == nil {
			clean = s.walk(last.CleanPrice, volatility)
			accrued = last.AccruedInterest
		}
		if err := s.store.Bonds.UpsertPrice(ctx, db.UpsertBondPriceParams{
			BondID: b.ID, PriceDate: today, CleanPrice: clean.Round(4), AccruedInterest: accrued.Round(4),
		}); err != nil {
			return err
		}
	}
	return nil
}

// walk applies a bounded random walk: value ± up to volatility% (US-030).
func (s *PriceGeneratorService) walk(value, volatilityPct decimal.Decimal) decimal.Decimal {
	s.mu.Lock()
	r := s.rng.Float64()*2 - 1 // [-1, 1)
	s.mu.Unlock()
	delta := value.Mul(volatilityPct.Div(decimal.NewFromInt(100))).Mul(decimal.NewFromFloat(r))
	out := value.Add(delta)
	if out.LessThanOrEqual(decimal.Zero) {
		return value
	}
	return out
}

func (s *PriceGeneratorService) today() time.Time {
	now := s.clock.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, Bangkok)
}

func stateView(st db.PriceGeneratorState) State {
	v := State{Enabled: st.Enabled, IntervalSeconds: st.IntervalSeconds, VolatilityPct: st.VolatilityPct.StringFixed(4)}
	if st.LastRunAt.Valid {
		v.LastRunAt = st.LastRunAt.Time.Format("2006-01-02T15:04:05-07:00")
	}
	return v
}
