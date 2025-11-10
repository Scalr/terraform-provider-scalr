package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/scalr/go-scalr"
	"github.com/scalr/go-scalr/v2/scalr/schemas"
)

var userElementType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"username":  types.StringType,
		"email":     types.StringType,
		"full_name": types.StringType,
	},
}

type userModel struct {
	Username types.String `tfsdk:"username"`
	Email    types.String `tfsdk:"email"`
	FullName types.String `tfsdk:"full_name"`
}

func userModelFromAPI(u *scalr.User) *userModel {
	return &userModel{
		Username: types.StringValue(u.Username),
		Email:    types.StringValue(u.Email),
		FullName: types.StringValue(u.FullName),
	}
}

func userModelFromAPIv2(u *schemas.User) *userModel {
	return &userModel{
		Username: types.StringValue(u.Attributes.Username),
		Email:    types.StringValue(u.Attributes.Email),
		FullName: types.StringPointerValue(u.Attributes.FullName),
	}
}
