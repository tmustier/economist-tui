package cmd

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"
	"github.com/tmustier/economist-tui/internal/rss"
)

var sectionsCmd = &cobra.Command{
	Use:   "sections",
	Short: "List available sections",
	Run:   runSections,
}

func runSections(cmd *cobra.Command, args []string) {
	fmt.Println("📚 Available sections:")
	fmt.Println()

	// Group aliases by canonical path
	pathToAliases := make(map[string][]string)
	for alias, path := range rss.Sections {
		pathToAliases[path] = append(pathToAliases[path], alias)
	}

	// Sort paths for consistent output
	var paths []string
	for path := range pathToAliases {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	for _, path := range paths {
		aliases := pathToAliases[path]
		sort.Strings(aliases)

		// Use shortest alias as primary name
		primary := shortestString(aliases)

		fmt.Printf("  %-20s", primary)

		// Show one alternate alias if available
		for _, a := range aliases {
			if a != primary {
				fmt.Printf(" (also: %s)", a)
				break
			}
		}
		fmt.Println()
	}

	fmt.Println()
	fmt.Println("Usage: economist headlines <section>")
}

func shortestString(strs []string) string {
	if len(strs) == 0 {
		return ""
	}
	shortest := strs[0]
	for _, s := range strs[1:] {
		if len(s) < len(shortest) {
			shortest = s
		}
	}
	return shortest
}
