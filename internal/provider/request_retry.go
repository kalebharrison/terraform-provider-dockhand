package provider

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultRequestRetryAttempts   = 6
	defaultRequestRetryMinDelay     = 500 * time.Millisecond
	defaultRequestRetryMaxDelay     = 5 * time.Second
	defaultRequestRetryMinSeconds   = 1
	defaultRequestRetryMaxSeconds   = 5
)

type requestRetryConfig struct {
	attempts int
	minDelay time.Duration
	maxDelay time.Duration
}

func defaultRequestRetryConfig() requestRetryConfig {
	return requestRetryConfig{
		attempts: defaultRequestRetryAttempts,
		minDelay: defaultRequestRetryMinDelay,
		maxDelay: defaultRequestRetryMaxDelay,
	}
}

func resolveRequestRetryConfig(config dockhandProviderModel) requestRetryConfig {
	retry := defaultRequestRetryConfig()

	if raw := strings.TrimSpace(os.Getenv("DOCKHAND_REQUEST_RETRY_ATTEMPTS")); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil && v > 0 {
			retry.attempts = v
		}
	}
	if !config.RequestRetryAttempts.IsNull() && !config.RequestRetryAttempts.IsUnknown() {
		if v := config.RequestRetryAttempts.ValueInt64(); v > 0 {
			retry.attempts = int(v)
		}
	}

	minSeconds := defaultRequestRetryMinSeconds
	if raw := strings.TrimSpace(os.Getenv("DOCKHAND_REQUEST_RETRY_MIN_SECONDS")); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil && v >= 0 {
			minSeconds = v
		}
	}
	if !config.RequestRetryMinSeconds.IsNull() && !config.RequestRetryMinSeconds.IsUnknown() {
		minSeconds = int(config.RequestRetryMinSeconds.ValueInt64())
	}

	maxSeconds := defaultRequestRetryMaxSeconds
	if raw := strings.TrimSpace(os.Getenv("DOCKHAND_REQUEST_RETRY_MAX_SECONDS")); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil && v > 0 {
			maxSeconds = v
		}
	}
	if !config.RequestRetryMaxSeconds.IsNull() && !config.RequestRetryMaxSeconds.IsUnknown() {
		maxSeconds = int(config.RequestRetryMaxSeconds.ValueInt64())
	}

	if minSeconds < 0 {
		minSeconds = 0
	}
	if maxSeconds < 1 {
		maxSeconds = defaultRequestRetryMaxSeconds
	}
	if minSeconds > maxSeconds {
		minSeconds = maxSeconds
	}

	retry.minDelay = time.Duration(minSeconds) * time.Second
	retry.maxDelay = time.Duration(maxSeconds) * time.Second
	return retry
}

func (c *requestRetryConfig) attemptsOrDefault() int {
	if c == nil || c.attempts < 1 {
		return defaultRequestRetryAttempts
	}
	return c.attempts
}

func (c *requestRetryConfig) sleep(ctx context.Context, attempt int) error {
	minDelay := defaultRequestRetryMinDelay
	maxDelay := defaultRequestRetryMaxDelay
	if c != nil {
		if c.minDelay > 0 {
			minDelay = c.minDelay
		}
		if c.maxDelay > 0 {
			maxDelay = c.maxDelay
		}
	}

	delay := minDelay
	for i := 0; i < attempt; i++ {
		delay *= 2
		if delay > maxDelay {
			delay = maxDelay
			break
		}
	}

	t := time.NewTimer(delay)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

func isTransientNetworkError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}

	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "connection refused"),
		strings.Contains(msg, "connection reset"),
		strings.Contains(msg, "broken pipe"),
		strings.Contains(msg, "no route to host"),
		strings.Contains(msg, "eof"),
		strings.Contains(msg, "timeout"),
		strings.Contains(msg, "temporarily unavailable"):
		return true
	}

	var ne net.Error
	if errors.As(err, &ne) && ne.Timeout() {
		return true
	}

	return false
}

func shouldRetryHTTPStatus(method string, status int) bool {
	switch method {
	case http.MethodGet, http.MethodDelete:
	default:
		return false
	}
	switch status {
	case http.StatusTooManyRequests, http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return true
	default:
		return false
	}
}

func shouldRetryRequest(method string, status int, err error) bool {
	if isTransientNetworkError(err) {
		return true
	}
	if err == nil {
		return shouldRetryHTTPStatus(method, status)
	}
	return extractHTTPStatusFromError(err) > 0 && shouldRetryHTTPStatus(method, extractHTTPStatusFromError(err))
}

func extractHTTPStatusFromError(err error) int {
	if err == nil {
		return 0
	}
	msg := err.Error()
	if !strings.Contains(msg, "status ") {
		return 0
	}
	var status int
	if _, scanErr := fmt.Sscanf(msg, "dockhand api returned status %d", &status); scanErr == nil {
		return status
	}
	if _, scanErr := fmt.Sscanf(msg, "dockhand login failed (status %d)", &status); scanErr == nil {
		return status
	}
	return 0
}
