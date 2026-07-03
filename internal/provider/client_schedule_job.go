package provider

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func (c *Client) ToggleSchedule(ctx context.Context, scheduleType string, id string, isSystem bool) (int, error) {
	path := "/api/schedules/" + url.PathEscape(scheduleType) + "/" + url.PathEscape(id) + "/toggle"
	if isSystem {
		path = "/api/schedules/system/" + url.PathEscape(id) + "/toggle"
	}
	return c.doJSONWithStatus(ctx, http.MethodPost, path, nil, nil, nil)
}

func (c *Client) RunSchedule(ctx context.Context, scheduleType string, id string) (int, error) {
	path := "/api/schedules/" + url.PathEscape(scheduleType) + "/" + url.PathEscape(id) + "/run"
	return c.doJSONWithStatus(ctx, http.MethodPost, path, nil, nil, nil)
}

func (c *Client) SubmitBatch(ctx context.Context, env string, entityType string, operation string, itemIDs []string) (*batchResponse, int, error) {
	items := make([]batchItemPayload, 0, len(itemIDs))
	for _, id := range itemIDs {
		trimmed := strings.TrimSpace(id)
		if trimmed == "" {
			continue
		}
		items = append(items, batchItemPayload{ID: trimmed})
	}
	if len(items) == 0 {
		return nil, 0, fmt.Errorf("at least one non-empty item id is required")
	}

	payload := batchRequestPayload{
		EntityType: strings.TrimSpace(entityType),
		Operation:  strings.TrimSpace(operation),
		Items:      items,
	}

	query := map[string]string{}
	if resolvedEnv := c.resolveEnv(env); resolvedEnv != "" {
		query["env"] = resolvedEnv
	}

	var raw map[string]any
	status, err := c.doJSONWithStatus(ctx, http.MethodPost, "/api/batch", query, payload, &raw)
	if err != nil {
		return nil, status, err
	}
	jobID := extractBatchJobID(raw)
	statusText := extractBatchStatus(raw)
	if jobID == "" && statusText == "" {
		return nil, status, fmt.Errorf("dockhand batch response missing jobId (response=%s)", mustJSON(raw))
	}
	return &batchResponse{
		JobID:  jobID,
		Status: statusText,
		Result: raw,
	}, status, nil
}

func (c *Client) GetJob(ctx context.Context, id string) (*jobResponse, int, error) {
	var raw map[string]any
	status, err := c.doJSONWithStatus(ctx, http.MethodGet, "/api/jobs/"+url.PathEscape(strings.TrimSpace(id)), nil, nil, &raw)
	if err != nil {
		return nil, status, err
	}
	out := parseJobResponse(raw)
	if strings.TrimSpace(out.ID) == "" {
		out.ID = strings.TrimSpace(id)
	}
	return &out, status, nil
}

func (c *Client) WaitForJob(ctx context.Context, id string, timeout time.Duration, pollInterval time.Duration) (*jobResponse, int, error) {
	if timeout <= 0 {
		timeout = 2 * time.Minute
	}
	if pollInterval <= 0 {
		pollInterval = 1 * time.Second
	}

	pollCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	for {
		out, status, err := c.GetJob(pollCtx, id)
		if err != nil {
			return nil, status, err
		}
		if isTerminalJobStatus(out.Status) {
			return out, status, nil
		}
		if err := sleepBackoffWithInterval(pollCtx, pollInterval); err != nil {
			return out, status, err
		}
	}
}

func (c *Client) GetSchedules(ctx context.Context) (*schedulesListResponse, int, error) {
	var out schedulesListResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodGet, "/api/schedules", nil, nil, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, nil
}

func (c *Client) GetScheduleExecutions(ctx context.Context, limit int64, offset int64) (*schedulesExecutionsResponse, int, error) {
	query := map[string]string{}
	if limit > 0 {
		query["limit"] = strconv.FormatInt(limit, 10)
	}
	if offset > 0 {
		query["offset"] = strconv.FormatInt(offset, 10)
	}

	var out schedulesExecutionsResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodGet, "/api/schedules/executions", query, nil, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, nil
}

func (c *Client) GetScheduleSettings(ctx context.Context) (*scheduleSettingsResponse, int, error) {
	var out scheduleSettingsResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodGet, "/api/schedules/settings", nil, nil, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, nil
}

func (c *Client) UpdateScheduleSettings(ctx context.Context, payload scheduleSettingsPayload) (*scheduleSettingsResponse, int, error) {
	var out scheduleSettingsResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodPut, "/api/schedules/settings", nil, payload, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, nil
}

func (c *Client) ReadScheduleStream(ctx context.Context, maxEvents int64, timeout time.Duration) ([]scheduleStreamEvent, int, error) {
	if maxEvents <= 0 {
		maxEvents = 1
	}
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	readCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	fullURL := c.baseURL.ResolveReference(&url.URL{Path: "/api/schedules/stream"}).String()
	req, err := http.NewRequestWithContext(readCtx, http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Accept", "text/event-stream")
	c.applyAuthHeaders(req)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer func() {
		_ = res.Body.Close()
	}()

	if res.StatusCode < 200 || res.StatusCode > 299 {
		body, _ := io.ReadAll(io.LimitReader(res.Body, 64<<10))
		if len(body) == 0 {
			return nil, res.StatusCode, fmt.Errorf("dockhand api returned status %d", res.StatusCode)
		}
		return nil, res.StatusCode, fmt.Errorf("dockhand api returned status %d: %s", res.StatusCode, strings.TrimSpace(string(body)))
	}

	events := make([]scheduleStreamEvent, 0, maxEvents)
	scanner := bufio.NewScanner(io.LimitReader(res.Body, 8<<20))
	scanner.Buffer(make([]byte, 64*1024), 8<<20)
	currentEvent := ""
	currentData := make([]string, 0, 4)

	flush := func() {
		if currentEvent == "" && len(currentData) == 0 {
			return
		}
		events = append(events, scheduleStreamEvent{
			Event: strings.TrimSpace(currentEvent),
			Data:  strings.TrimSpace(strings.Join(currentData, "\n")),
		})
		currentEvent = ""
		currentData = currentData[:0]
	}

	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimRight(line, "\r")
		if line == "" {
			flush()
			if int64(len(events)) >= maxEvents {
				break
			}
			continue
		}

		switch {
		case strings.HasPrefix(line, "event:"):
			currentEvent = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
		case strings.HasPrefix(line, "data:"):
			currentData = append(currentData, strings.TrimSpace(strings.TrimPrefix(line, "data:")))
		}
	}

	flush()

	if err := scanner.Err(); err != nil {
		if len(events) > 0 {
			return events, res.StatusCode, nil
		}
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(readCtx.Err(), context.DeadlineExceeded) {
			return []scheduleStreamEvent{}, res.StatusCode, nil
		}
		return nil, res.StatusCode, err
	}

	if events == nil {
		events = []scheduleStreamEvent{}
	}
	return events, res.StatusCode, nil
}

func isTerminalJobStatus(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "done", "failed", "error", "cancelled", "canceled", "complete", "completed", "success":
		return true
	default:
		return false
	}
}

func extractBatchStatus(payload map[string]any) string {
	if payload == nil {
		return ""
	}

	if status := strings.TrimSpace(firstString(payload, "status", "state", "type")); status != "" {
		switch strings.ToLower(status) {
		case "complete", "completed", "success", "ok":
			return "done"
		case "queued":
			return "queued"
		case "pending":
			return "pending"
		case "running", "processing":
			return "running"
		case "failed", "failure", "error":
			return "failed"
		case "cancelled", "canceled":
			return "cancelled"
		default:
			return status
		}
	}

	for _, key := range []string{"job", "data", "result"} {
		if m := firstMap(payload, key); m != nil {
			if status := extractBatchStatus(m); status != "" {
				return status
			}
		}
	}

	return ""
}

func parseJobLines(value any) []jobLineResponse {
	raw, ok := value.([]any)
	if !ok {
		return nil
	}

	lines := make([]jobLineResponse, 0, len(raw))
	for _, entry := range raw {
		var data map[string]any
		switch v := entry.(type) {
		case map[string]any:
			data = v
		case string:
			data = map[string]any{"message": v}
		default:
			data = map[string]any{"value": v}
		}
		lines = append(lines, jobLineResponse{Data: data})
	}
	return lines
}
