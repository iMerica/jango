package rest

import (
	"fmt"
	"math"
	"net"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/iMerica/jango/cache"
)

var defaultThrottleCache cache.Cache = cache.NewMemoryCache()

type throttleHistory struct {
	Requests []time.Time
}

type RateThrottle struct {
	Rate  string
	Cache cache.Cache
	Now   func() time.Time
	Scope string
}

type AnonRateThrottle struct {
	RateThrottle
}

type UserRateThrottle struct {
	RateThrottle
}

type ScopedRateThrottle struct {
	RateThrottle
	Rates map[string]string
}

func ParseRate(rate string) (int, time.Duration, error) {
	parts := strings.Split(strings.TrimSpace(rate), "/")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid throttle rate %q", rate)
	}
	limit, err := strconv.Atoi(parts[0])
	if err != nil || limit <= 0 {
		return 0, 0, fmt.Errorf("invalid throttle rate %q", rate)
	}
	switch strings.ToLower(strings.TrimSpace(parts[1])) {
	case "s", "sec", "second", "seconds":
		return limit, time.Second, nil
	case "m", "min", "minute", "minutes":
		return limit, time.Minute, nil
	case "h", "hour", "hours":
		return limit, time.Hour, nil
	case "d", "day", "days":
		return limit, 24 * time.Hour, nil
	default:
		return 0, 0, fmt.Errorf("invalid throttle period %q", parts[1])
	}
}

func (t AnonRateThrottle) AllowRequest(req *APIRequest, view interface{}) bool {
	if userIsAuthenticated(req.User) {
		return true
	}
	return t.allow(req, "anon", requestIdentity(req), t.Rate)
}

func (t AnonRateThrottle) Wait(req *APIRequest, view interface{}) int {
	return t.wait(req, "anon", requestIdentity(req), t.Rate)
}

func (t UserRateThrottle) AllowRequest(req *APIRequest, view interface{}) bool {
	if !userIsAuthenticated(req.User) {
		return true
	}
	return t.allow(req, "user", userIdentity(req.User), t.Rate)
}

func (t UserRateThrottle) Wait(req *APIRequest, view interface{}) int {
	if !userIsAuthenticated(req.User) {
		return 0
	}
	return t.wait(req, "user", userIdentity(req.User), t.Rate)
}

func (t ScopedRateThrottle) AllowRequest(req *APIRequest, view interface{}) bool {
	scope := throttleScope(view)
	if scope == "" {
		scope = t.Scope
	}
	if scope == "" {
		return true
	}
	rate := t.Rate
	if rate == "" && t.Rates != nil {
		rate = t.Rates[scope]
	}
	if rate == "" {
		return true
	}
	return t.allow(req, "scoped:"+scope, throttleActorIdentity(req), rate)
}

func (t ScopedRateThrottle) Wait(req *APIRequest, view interface{}) int {
	scope := throttleScope(view)
	if scope == "" {
		scope = t.Scope
	}
	if scope == "" {
		return 0
	}
	rate := t.Rate
	if rate == "" && t.Rates != nil {
		rate = t.Rates[scope]
	}
	if rate == "" {
		return 0
	}
	return t.wait(req, "scoped:"+scope, throttleActorIdentity(req), rate)
}

func (t RateThrottle) allow(req *APIRequest, bucket, identity, rate string) bool {
	limit, window, err := ParseRate(rate)
	if err != nil || identity == "" {
		return true
	}
	now := t.now()
	key := throttleKey(bucket, identity)
	history := t.history(key)
	requests := trimThrottleHistory(history.Requests, now, window)
	if len(requests) >= limit {
		t.cache().Set(key, throttleHistory{Requests: requests}, window)
		return false
	}
	requests = append(requests, now)
	t.cache().Set(key, throttleHistory{Requests: requests}, window)
	return true
}

func (t RateThrottle) wait(req *APIRequest, bucket, identity, rate string) int {
	limit, window, err := ParseRate(rate)
	if err != nil || identity == "" {
		return 0
	}
	now := t.now()
	history := t.history(throttleKey(bucket, identity))
	requests := trimThrottleHistory(history.Requests, now, window)
	if len(requests) < limit || len(requests) == 0 {
		return 0
	}
	remaining := window - now.Sub(requests[0])
	if remaining <= 0 {
		return 0
	}
	return int(math.Ceil(remaining.Seconds()))
}

func (t RateThrottle) history(key string) throttleHistory {
	value, ok := t.cache().Get(key)
	if !ok {
		return throttleHistory{}
	}
	switch history := value.(type) {
	case throttleHistory:
		return history
	case []time.Time:
		return throttleHistory{Requests: history}
	default:
		return throttleHistory{}
	}
}

func (t RateThrottle) cache() cache.Cache {
	if t.Cache != nil {
		return t.Cache
	}
	return defaultThrottleCache
}

func (t RateThrottle) now() time.Time {
	if t.Now != nil {
		return t.Now()
	}
	return time.Now()
}

func trimThrottleHistory(requests []time.Time, now time.Time, window time.Duration) []time.Time {
	trimmed := requests[:0]
	for _, requestTime := range requests {
		if now.Sub(requestTime) < window {
			trimmed = append(trimmed, requestTime)
		}
	}
	return trimmed
}

func throttleKey(bucket, identity string) string {
	return "rest:throttle:" + bucket + ":" + identity
}

func requestIdentity(req *APIRequest) string {
	if forwarded := strings.TrimSpace(req.Header.Get("X-Forwarded-For")); forwarded != "" {
		return strings.TrimSpace(strings.Split(forwarded, ",")[0])
	}
	host, _, err := net.SplitHostPort(req.RemoteAddr)
	if err == nil && host != "" {
		return host
	}
	return req.RemoteAddr
}

func throttleActorIdentity(req *APIRequest) string {
	if userIsAuthenticated(req.User) {
		return "user:" + userIdentity(req.User)
	}
	return "anon:" + requestIdentity(req)
}

func userIdentity(user interface{}) string {
	if user == nil {
		return ""
	}
	if pk, ok := user.(interface{ PKValue() interface{} }); ok {
		return fmt.Sprint(pk.PKValue())
	}
	value := reflect.ValueOf(user)
	for value.Kind() == reflect.Ptr {
		if value.IsNil() {
			return ""
		}
		value = value.Elem()
	}
	if value.Kind() == reflect.Struct {
		field := value.FieldByName("ID")
		if field.IsValid() && field.CanInterface() {
			return fmt.Sprint(field.Interface())
		}
	}
	return fmt.Sprint(user)
}

func throttleScope(view interface{}) string {
	switch v := view.(type) {
	case interface{ GetThrottleScope() string }:
		return v.GetThrottleScope()
	case APIView:
		return v.ThrottleScope
	default:
		value := reflect.ValueOf(view)
		for value.Kind() == reflect.Ptr {
			if value.IsNil() {
				return ""
			}
			value = value.Elem()
		}
		if value.Kind() == reflect.Struct {
			field := value.FieldByName("ThrottleScope")
			if field.IsValid() && field.Kind() == reflect.String {
				return field.String()
			}
		}
	}
	return ""
}

func (v APIView) GetThrottleScope() string {
	return v.ThrottleScope
}

func (v APIView) versioningStrategy() VersioningStrategy {
	return v.Versioning
}
