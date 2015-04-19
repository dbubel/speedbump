package speedbump

import (
	"os"
	"testing"
	"time"

	"github.com/facebookgo/clock"
	"github.com/stretchr/testify/assert"

	"gopkg.in/redis.v2"
)

func createClient() *redis.Client {
	if os.Getenv("WERCKER_REDIS_HOST") != "" {
		return redis.NewTCPClient(&redis.Options{
			Addr:     os.Getenv("WERCKER_REDIS_HOST") + ":6379",
			Password: "",
			DB:       0,
		})
	}

	return redis.NewTCPClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
}

func Test_NewLimiter(t *testing.T) {
	client := createClient()
	hasher := PerSecondHasher{}

	NewLimiter(client, hasher, 10)
}

func Test_Attempts(t *testing.T) {
	client := createClient()
	mock := clock.NewMock()
	hasher := PerSecondHasher{
		Clock: mock,
	}

	limiter := NewLimiter(client, hasher, 5)

	has, err := limiter.Has("127.0.0.1")
	assert.Nil(t, err)
	assert.False(t, has)

	ok, err := limiter.Attempt("127.0.0.1")
	assert.Nil(t, err)
	assert.True(t, ok)

	limiter.Attempt("127.0.0.1")
	limiter.Attempt("127.0.0.1")
	limiter.Attempt("127.0.0.1")
	limiter.Attempt("127.0.0.1")
	limiter.Attempt("127.0.0.1")
	limiter.Attempt("127.0.0.1")
	limiter.Attempt("127.0.0.2")
	limiter.Attempt("127.0.0.1")
	limiter.Attempt("127.0.0.2")
	ok, err = limiter.Attempt("127.0.0.1")

	assert.Nil(t, err)
	assert.False(t, ok)

	left, err := limiter.Left("127.0.0.1")
	assert.Nil(t, err)
	assert.Equal(t, 0, left)

	left, err = limiter.Left("127.0.0.2")
	assert.Nil(t, err)
	assert.True(t, left > 0)

	ok, err = limiter.Attempt("127.0.0.2")
	assert.Nil(t, err)
	assert.True(t, ok)

	mock.Add(time.Second)

	left, err = limiter.Left("127.0.0.1")
	assert.Nil(t, err)
	assert.Equal(t, 5, left)
}
