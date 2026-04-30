package report

import (
	"errors"
	"testing"

	"github.com/1340691923/xwl_bi/engine/logs"
	"github.com/1340691923/xwl_bi/platform-basic-libs/service/myapp"
	"github.com/garyburd/redigo/redis"
	"go.uber.org/zap"
)

func TestTableIDResolverResolveUsesCache(t *testing.T) {
	logs.Logger = zap.NewNop()

	callCount := 0
	resolver := NewTableIDResolver(func(key string) (string, error) {
		callCount++
		if key != myapp.BuildAppTableIDKey("1001", "demo") {
			t.Fatalf("unexpected key %q", key)
		}
		return "51", nil
	})

	first, err := resolver.Resolve("1001", "demo")
	if err != nil {
		t.Fatalf("first Resolve returned error: %v", err)
	}
	second, err := resolver.Resolve("1001", "demo")
	if err != nil {
		t.Fatalf("second Resolve returned error: %v", err)
	}

	if first != "51" || second != "51" {
		t.Fatalf("Resolve returned %q and %q, want both %q", first, second, "51")
	}
	if callCount != 1 {
		t.Fatalf("fetch call count = %d, want %d", callCount, 1)
	}
}

func TestTableIDResolverResolveMapsErrors(t *testing.T) {
	logs.Logger = zap.NewNop()

	nilResolver := NewTableIDResolver(func(key string) (string, error) {
		return "", redis.ErrNil
	})
	if _, err := nilResolver.Resolve("1001", "demo"); err == nil || err.Error() != ERROR_TABLE[AppParmasErr] {
		t.Fatalf("Resolve redis nil error = %v, want %q", err, ERROR_TABLE[AppParmasErr])
	}

	serverResolver := NewTableIDResolver(func(key string) (string, error) {
		return "", errors.New("redis down")
	})
	if _, err := serverResolver.Resolve("1001", "demo"); err == nil || err.Error() != ERROR_TABLE[ServerErr] {
		t.Fatalf("Resolve server error = %v, want %q", err, ERROR_TABLE[ServerErr])
	}
}

func TestTableIDResolverInvalidate(t *testing.T) {
	logs.Logger = zap.NewNop()

	resolver := NewTableIDResolver(func(key string) (string, error) {
		return "51", nil
	})
	cacheKey := myapp.BuildAppTableIDKey("1001", "demo")
	resolver.cache.Store(cacheKey, "51")

	resolver.Invalidate("1001", "demo")
	if _, ok := resolver.cache.Load(cacheKey); ok {
		t.Fatal("cache entry was not invalidated by appid/appkey")
	}

	resolver.cache.Store(cacheKey, "51")
	resolver.InvalidateByKey(cacheKey)
	if _, ok := resolver.cache.Load(cacheKey); ok {
		t.Fatal("cache entry was not invalidated by cache key")
	}
}

func TestDefaultTableIDResolverReceivesSideCacheInvalidation(t *testing.T) {
	logs.Logger = zap.NewNop()

	resolver := DefaultTableIDResolver()
	resolver.Clear()

	cacheKey := myapp.BuildAppTableIDKey("1001", "demo")
	resolver.cache.Store(cacheKey, "51")

	myapp.NotifyAppTableIDCacheChanged(cacheKey)
	if _, ok := resolver.cache.Load(cacheKey); ok {
		t.Fatal("default resolver cache entry was not invalidated by side cache notification")
	}
}
