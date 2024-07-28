package cmdimpl

import (
	"context"
	"fmt"

	"github.com/Ruixi-rebirth/mygo/internal/output"
	"github.com/Ruixi-rebirth/mygo/pkg/registry"
)

func SearchPackages(ctx context.Context, query string, limit int) error {
	if limit <= 0 {
		limit = 10
	}

	client := registry.NewClient()

	output.Infof("Searching GitHub for %q...", output.Bold(query))

	repos, err := client.SearchGitHub(query, limit, "stars")
	if err != nil {
		return fmt.Errorf("github search: %w", err)
	}

	if len(repos) == 0 {
		output.Warnf("No Go repositories found for %q", query)
		return nil
	}

	fmt.Println()
	t := output.NewTable("Repository", "Stars", "Description")
	for _, r := range repos {
		desc := r.Description
		if len(desc) > 55 {
			desc = desc[:52] + "..."
		}
		if desc == "" {
			desc = output.Dim("(no description)")
		}
		t.Row(
			output.Cyan(r.FullName),
			fmt.Sprintf("★ %d", r.Stars),
			desc,
		)
	}
	t.Print()
	fmt.Println()
	output.Printf(output.Info, "%s %d repos on GitHub (sorted by stars)", output.Dim("──"), len(repos))
	return nil
}
