package provider

import (
	"strconv"

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
