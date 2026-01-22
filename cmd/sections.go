package cmd

import (
	"fmt"

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

	sections := rss.SectionList()
	for _, section := range sections {
		primary := section.Primary
		fmt.Printf("  %-20s", primary)

		// Show one alternate alias if available
		for _, a := range section.Aliases {
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
