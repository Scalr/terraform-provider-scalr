package framework

import (
	"github.com/scalr/go-scalr"
	scalrV2 "github.com/scalr/go-scalr/v2/scalr"
)

type Clients struct {
	Client   *scalr.Client
	ClientV2 *scalrV2.Client
}
