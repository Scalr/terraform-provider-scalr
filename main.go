package main

import (
	"github.com/hashicorp/terraform/plugin"
	"github.com/scalr/terraform-provider-scalr/scalr"
)
import "fmt"

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: scalr.Provider})
	greeting := "Bob"
	fmt.Printf("Hello, %s", greeting)
}
