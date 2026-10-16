package domain

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

func TestStatusesMatchMigrationChecks(t *testing.T) {
	migration, err := os.ReadFile("../../migrations/transaction/0001_init_transaction_history.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	statusLists := regexp.MustCompile(`CHECK \((?:from_status IS NULL OR )?(?:status|to_status|from_status) IN \(([^)]*)\)\)`).FindAllStringSubmatch(string(migration), -1)
	if len(statusLists) != 3 {
		t.Fatalf("found %d status IN lists, want header and both event checks", len(statusLists))
	}
	want := make([]string, len(Statuses))
	for index, status := range Statuses {
		want[index] = "'" + status + "'"
	}
	wantList := strings.Join(want, ",")
	for index, match := range statusLists {
		got := strings.ReplaceAll(match[1], " ", "")
		if got != wantList {
			t.Errorf("status IN list %d = %q, want %q", index+1, got, wantList)
		}
	}
}
