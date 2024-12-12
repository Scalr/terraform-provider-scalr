package main

import (
	"flag"
	"log"

	"github.com/scalr/terraform-provider-scalr/skaff"
)

func main() {
	t := flag.String("type", "", "Type of scaffolding to create.")
	n := flag.String(
		"name", "", "A name in snake case as it will appear in configuration (e.g., agent_pool).",
	)
	flag.Parse()

	err := skaff.Generate(*t, *n)
	if err != nil {
		log.Fatal(err)
	}
}
