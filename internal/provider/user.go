package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scalr/go-scalr"
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
