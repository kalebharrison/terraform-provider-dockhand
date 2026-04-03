package provider

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func stringValueOrNull(v *string) types.String {
	if v == nil {
		return types.StringNull()
	}
	return types.StringValue(*v)
}

func int64StringValueOrNull(v *int64) types.String {
	if v == nil {
		return types.StringNull()
	}
	return types.StringValue(strconv.FormatInt(*v, 10))
}

func optionalStringValue(v types.String) *string {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	value := strings.TrimSpace(v.ValueString())
	if value == "" {
		return nil
	}
	return &value
}

func optionalInt64StringValue(name string, v types.String) (*int64, error) {
	if v.IsNull() || v.IsUnknown() {
		return nil, nil
	}
	raw := strings.TrimSpace(v.ValueString())
	if raw == "" {
		return nil, nil
	}
	parsed, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("%s must be a numeric string: %w", name, err)
	}
	return &parsed, nil
}

func sortedMapKeys(v map[string]any) []string {
	if len(v) == 0 {
		return []string{}
	}
	keys := make([]string, 0, len(v))
	for key := range v {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
