package classify

import (
	"testing"

	"github.com/9sako6/craftcompass/internal/config"
)

func TestCommandNameSkipsEnvAssignmentsAndWrappers(t *testing.T) {
	t.Parallel()

	got := CommandName(`FOO=1 BAR=2 noglob go test ./...`)
	if got != "go" {
		t.Fatalf("expected command name go, got %q", got)
	}
}

func TestCategoryPrefersSpecificGoTestRule(t *testing.T) {
	t.Parallel()

	cfg := config.Default()
	got := Category(`go test ./...`, cfg.Categories)
	if got != "test" {
		t.Fatalf("expected go test to map to test, got %q", got)
	}
}

func TestCategoryFallsBackToMisc(t *testing.T) {
	t.Parallel()

	cfg := config.Default()
	got := Category(`unknown-tool --flag`, cfg.Categories)
	if got != "misc" {
		t.Fatalf("expected fallback misc, got %q", got)
	}
}
