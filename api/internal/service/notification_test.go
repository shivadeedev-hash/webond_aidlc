package service

import (
	"context"
	"testing"

	"github.com/shopspring/decimal"

	"github.com/webond/api/internal/db"
)

// TestNotification_EmitHonorsEmailPreference writes an in-app row always and an
// email row only when email is enabled (notification.go Emit/emit; US-029, FR-24).
func TestNotification_EmitHonorsEmailPreference(t *testing.T) {
	store := openTestStore(t)
	pool := rawPool(t)
	clock := testClock()
	ns := NewNotificationService(store, clock)
	ctx := context.Background()

	userID := seedActiveUser(t, store, clock, decimal.Zero, 3)
	t.Cleanup(func() { cleanupUser(t, pool, userID) })

	// Email disabled: only the in-app row is written.
	if err := ns.Emit(ctx, Notice{
		UserID: userID, Category: db.NotificationOutboxCategoryOrder,
		Subject: "Order filled", Body: "Your order completed", EmailEnabled: false,
	}); err != nil {
		t.Fatalf("emit (email off): %v", err)
	}
	inApp, _, err := ns.ListInApp(ctx, userID, 1, 50)
	if err != nil {
		t.Fatalf("list in-app: %v", err)
	}
	if len(inApp) != 1 {
		t.Fatalf("expected 1 in-app row, got %d", len(inApp))
	}
	inbox, _, err := ns.ListInbox(ctx, userID, 1, 50)
	if err != nil {
		t.Fatalf("list inbox: %v", err)
	}
	if len(inbox) != 0 {
		t.Fatalf("expected 0 email rows when email disabled, got %d", len(inbox))
	}

	// Email enabled: both an in-app and an email row are written.
	if err := ns.Emit(ctx, Notice{
		UserID: userID, Category: db.NotificationOutboxCategoryCoupon,
		Subject: "Coupon", Body: "Coupon received", EmailEnabled: true,
	}); err != nil {
		t.Fatalf("emit (email on): %v", err)
	}
	inApp2, _, _ := ns.ListInApp(ctx, userID, 1, 50)
	inbox2, _, _ := ns.ListInbox(ctx, userID, 1, 50)
	if len(inApp2) != 2 {
		t.Fatalf("expected 2 in-app rows, got %d", len(inApp2))
	}
	if len(inbox2) != 1 {
		t.Fatalf("expected 1 email row when enabled, got %d", len(inbox2))
	}
}

// TestNotification_TokenEmailAndMarkRead covers token-carrying email-only rows
// (always written) and MarkRead ownership (notification.go emit/MarkRead;
// US-002/004/029).
func TestNotification_TokenEmailAndMarkRead(t *testing.T) {
	store := openTestStore(t)
	pool := rawPool(t)
	clock := testClock()
	ns := NewNotificationService(store, clock)
	ctx := context.Background()

	userID := seedActiveUser(t, store, clock, decimal.Zero, 3)
	t.Cleanup(func() { cleanupUser(t, pool, userID) })

	// EmailOnly + token: suppress the in-app row, always write the email row
	// even though EmailEnabled is false (verify/reset must land in the inbox).
	if err := ns.Emit(ctx, Notice{
		UserID: userID, Category: db.NotificationOutboxCategoryVerify,
		Subject: "Verify", Body: "Click to verify", Token: "verify-token-xyz",
		EmailEnabled: false, EmailOnly: true,
	}); err != nil {
		t.Fatalf("emit token email: %v", err)
	}
	inApp, _, _ := ns.ListInApp(ctx, userID, 1, 50)
	if len(inApp) != 0 {
		t.Fatalf("EmailOnly should suppress in-app rows, got %d", len(inApp))
	}
	inbox, _, _ := ns.ListInbox(ctx, userID, 1, 50)
	if len(inbox) != 1 || !inbox[0].Token.Valid || inbox[0].Token.String != "verify-token-xyz" {
		t.Fatalf("token email row missing/incorrect: %+v", inbox)
	}

	// Now write a normal in-app notice and mark it read.
	if err := ns.Emit(ctx, Notice{
		UserID: userID, Category: db.NotificationOutboxCategoryGeneral,
		Subject: "Hello", Body: "World", EmailEnabled: false,
	}); err != nil {
		t.Fatalf("emit general: %v", err)
	}
	rows, _, _ := ns.ListInApp(ctx, userID, 1, 50)
	if len(rows) == 0 {
		t.Fatal("expected an in-app row to mark read")
	}
	if err := ns.MarkRead(ctx, userID, rows[0].ID); err != nil {
		t.Fatalf("mark read: %v", err)
	}
}
