// internal/cache/redis_session_store_test.go
package cache_test

import (
	"context"
	"github.com/go-redis/redismock/v9"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"user-service/internal/cache"
)

func TestRedisSessionStore(t *testing.T) {
	client, mock := redismock.NewClientMock()
	store := cache.NewSessionStore(client, time.Minute)

	// BlacklistToken
	mock.ExpectSet("tok", "1", time.Minute).SetVal("OK")
	assert.NoError(t, store.BlacklistToken(context.Background(), "tok"))

	// IsBlacklisted true
	mock.ExpectExists("tok").SetVal(1)
	ok, err := store.IsBlacklisted(context.Background(), "tok")
	assert.NoError(t, err)
	assert.True(t, ok)

	// IsBlacklisted false
	mock.ExpectExists("other").SetVal(0)
	ok, err = store.IsBlacklisted(context.Background(), "other")
	assert.NoError(t, err)
	assert.False(t, ok)

	assert.NoError(t, mock.ExpectationsWereMet())
}
