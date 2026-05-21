package framework

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/scalr/go-scalr"
	scalrV2 "github.com/scalr/go-scalr/v2/scalr"
)

type Clients struct {
	Client   *scalr.Client
	ClientV2 *scalrV2.Client
}

type AttrGetter interface {
	GetAttribute(ctx context.Context, path path.Path, target interface{}) diag.Diagnostics
}
