package provider

import "github.com/hashicorp/terraform-plugin-framework/types"

func stringValueOrNull(v *string) types.String {
	if v == nil {
		return types.StringNull()
	}
	return types.StringValue(*v)
}
