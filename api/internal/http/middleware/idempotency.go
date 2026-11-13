package middleware

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/webond/api/internal/domain"
)

// IdempotencyHeader is the header carrying the client-supplied key (api-spec.md
// § Idempotency; EX-11, NFR-05).
const IdempotencyHeader = "Idempotency-Key"

// ContextIdempotencyKey is where the validated key is stored for the handler,
// which persists it on the order row (orders.idempotency_key).
const ContextIdempotencyKey = "idempotencyKey"

// IdempotencyStore records in-flight/seen keys. The DB unique constraint on
// orders.idempotency_key is the durable guarantee; this middleware adds a fast
// in-process guard against rapid duplicate submissions. A production multi-node
// deploy would back this with a shared store (Redis) — see components.md.
type IdempotencyStore interface {
	// Reserve records the key and returns false if it was already present.
	Reserve(key string) bool
}

// memoryIdempotencyStore is a TTL-bounded in-memory key set.
type memoryIdempotencyStore struct {
	mu   sync.Mutex
	seen map[string]time.Time
	ttl  time.Duration
	now  func() time.Time
}

// NewMemoryIdempotencyStore builds an in-memory store keeping keys for ttl.
func NewMemoryIdempotencyStore(ttl time.Duration) IdempotencyStore {
	return &memoryIdempotencyStore{
		seen: make(map[string]time.Time),
		ttl:  ttl,
		now:  time.Now,
	}
}

func (s *memoryIdempotencyStore) Reserve(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.now()
	// Opportunistic cleanup of expired keys.
	for k, at := range s.seen {
		if now.Sub(at) > s.ttl {
			delete(s.seen, k)
		}
	}
	if _, exists := s.seen[key]; exists {
		return false
	}
	s.seen[key] = now
	return true
}

// Idempotency requires and validates the Idempotency-Key header on the routes
// it guards (order POSTs). It stores the key in the context for the handler and
// rejects an immediate duplicate with 409. Missing key → 400 VALIDATION_ERROR.
func Idempotency(store IdempotencyStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetHeader(IdempotencyHeader)
		if key == "" {
			abortError(c, domain.CodeValidationError, "Idempotency-Key header is required")
			return
		}
		if !store.Reserve(key) {
			abortError(c, domain.CodeCancelNotAllowed, "Duplicate request for this idempotency key")
			return
		}
		c.Set(ContextIdempotencyKey, key)
		c.Next()
	}
}
