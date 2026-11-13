package service

import (
	"context"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/shopspring/decimal"

	"github.com/webond/api/internal/db"
)

// TestProperty_OrderAtomicity validates correctness.md P3: an order is a single
// DB transaction — on failure nothing changes (cash + holding unchanged), and
// on success cash leaves the wallet exactly equal to the cost added to the
// holding (value conservation). It drives random buy amounts against a seeded
// fund and asserts the post-state matches the accepted/rejected outcome.
func TestProperty_OrderAtomicity(t *testing.T) {
	store := openTestStore(t)
	pool := rawPool(t)
	clock := testClock()
	ws := NewWalletService(store, clock)
	ts := NewTradingService(store, ws, NewNotificationService(store, clock), clock)
	ctx := context.Background()

	params := gopter.DefaultTestParameters()
	params.MinSuccessfulTests = 40
	properties := gopter.NewProperties(params)

	properties.Property("buy is atomic and value-conserving", prop.ForAll(
		func(amountInt int) bool {
			startBalance := decimal.RequireFromString("50000")
			userID := seedActiveUser(t, store, clock, startBalance, 5)
			fundID := seedFund(t, store, clock, "PT"+shortID(), decimal.RequireFromString("12.5"), 3)
			defer func() {
				cleanupUser(t, pool, userID)
				cleanupAsset(t, pool, "fund", fundID)
			}()

			amount := decimal.NewFromInt(int64(amountInt))
			before, _ := ws.Get(ctx, userID)

			_, err := ts.PlaceOrder(ctx, OrderInput{
				UserID: userID, AssetType: "fund", AssetID: fundID, Side: "buy",
				OrderMode: "amount", Amount: amount,
			})

			after, _ := ws.Get(ctx, userID)
			h, herr := store.Holdings.Get(ctx, db.GetHoldingParams{
				UserID: userID, AssetType: db.HoldingsAssetTypeFund, AssetID: fundID,
			})

			accepted := err == nil
			if accepted {
				// Cash dropped by exactly `amount`; holding total_cost == amount.
				if !after.Balance.Equal(before.Balance.Sub(amount).Round(4)) {
					return false
				}
				if herr != nil || !h.TotalCost.Equal(amount.Round(4)) {
					return false
				}
				// No cash left reserved.
				if !after.Reserved.Equal(decimal.Zero.Round(4)) {
					return false
				}
			} else {
				// Rejected (below min-purchase 500, or > available): nothing changed.
				if !after.Balance.Equal(before.Balance) || !after.Reserved.Equal(before.Reserved) {
					return false
				}
				if herr == nil { // holding must not exist
					return false
				}
			}
			return true
		},
		gen.IntRange(100, 60000),
	))

	properties.TestingRun(t)
}
