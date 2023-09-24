//go:build tools

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	var dir string
	flag.StringVar(&dir, "dir", "", "The directory to search for the *.md files.")
	flag.Parse()

	if dir == "" {
		log.Fatal("Directory not specified.")
	}

	fmt.Printf("--- Ordering the documentation files in %q:\n", dir)

	pattern := strings.TrimSuffix(dir, "/") + "/*.md"
	files, err := filepath.Glob(pattern)
	if err != nil {
		log.Fatal(err)
	}

	order := 1
	for _, f := range files {
		fmt.Printf("> Processing %q, order = %d...\n", filepath.Base(f), order)

		bytes, err := os.ReadFile(f)
		if err != nil {
			log.Fatalln(err)
		}

		lines := strings.Split(string(bytes), "\n")
		frontMatterFound := false

		for j, l := range lines {
			if l == "---" {
				if frontMatterFound {
					// This is the closing Front Matter delimiter.
					// No `order: ` records were found,
					// so we insert a new one.
					lines = append(lines[:j+1], lines[j:]...)
					lines[j] = fmt.Sprintf("order: %d", order)
					order++
					fmt.Println("  * Added order record.")
					break
				} else {
					frontMatterFound = true
					continue
				}
			}
			if frontMatterFound && strings.HasPrefix(l, "order:") {
				lines[j] = fmt.Sprintf("order: %d", order)
				order++
				fmt.Println("  * Updated order record.")
				break
			}
		}
		if frontMatterFound {
			err = os.WriteFile(f, []byte(strings.Join(lines, "\n")), 0644)
			if err != nil {
				log.Fatalln(err)
			}
			fmt.Println("  * File saved.")
		} else {
			fmt.Println("  ? No Front Matter block detected, skipping.")
		}
	}

	fmt.Println("--- Done.")
}
