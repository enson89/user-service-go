// internal/cache/redis_session_store_test.go
package cache_test

import (
	"testing"
	"time"

	"github.com/go-redis/redismock/v9"

	"github.com/stretchr/testify/assert"

	"github.com/enson89/user-service-go/internal/cache"
)

func TestRedisSessionStore(t *testing.T) {
	client, mock := redismock.NewClientMock()
	store := cache.NewSessionStore(client, time.Minute)

	// BlacklistToken
	mock.ExpectSet("tok", "1", time.Minute).SetVal("OK")
	assert.NoError(t, store.BlacklistToken(t.Context(), "tok"))

	// IsBlacklisted true
	mock.ExpectExists("tok").SetVal(1)
	ok, err := store.IsBlacklisted(t.Context(), "tok")
	assert.NoError(t, err)
	assert.True(t, ok)

	// IsBlacklisted false
	mock.ExpectExists("other").SetVal(0)
	ok, err = store.IsBlacklisted(t.Context(), "other")
	assert.NoError(t, err)
	assert.False(t, ok)

	assert.NoError(t, mock.ExpectationsWereMet())
}
