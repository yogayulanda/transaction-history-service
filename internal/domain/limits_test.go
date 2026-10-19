package domain

import (
	"os"
	"regexp"
	"strconv"
	"testing"
)

func TestPayloadLimitsMatchEventDocAndMigration(t *testing.T) {
	doc, err := os.ReadFile("../../docs/events/transaction-status-v1.md")
	if err != nil {
		t.Fatal(err)
	}
	migration, err := os.ReadFile("../../migrations/transaction/0001_init_transaction_history.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	columnSizes := map[string]int{}
	for _, match := range regexp.MustCompile(`(?m)^\s+(\w+)\s+N?VARCHAR\((\d+)\)`).FindAllStringSubmatch(string(migration), -1) {
		if _, seen := columnSizes[match[1]]; !seen {
			columnSizes[match[1]], _ = strconv.Atoi(match[2])
		}
	}
	for field, limit := range PayloadMaxLengths {
		row := regexp.MustCompile("(?m)^\\| `" + field + "` \\|[^|]*\\|[^\\n]*max (\\d+) characters").FindStringSubmatch(string(doc))
		if row == nil {
			t.Errorf("%s: no single-field doc row with a max length", field)
			continue
		}
		if documented, _ := strconv.Atoi(row[1]); documented != limit {
			t.Errorf("%s: doc limit %d, code limit %d", field, documented, limit)
		}
		if column, ok := columnSizes[field]; !ok || column != limit {
			t.Errorf("%s: migration column size %d (found=%v), code limit %d", field, column, ok, limit)
		}
	}
}

func TestStatusEventInputWithinLimits(t *testing.T) {
	if !(StatusEventInput{AppName: "app_a"}).WithinLimits() {
		t.Fatal("valid input rejected")
	}
	if (StatusEventInput{AppName: "12345678901234567"}).WithinLimits() {
		t.Fatal("app_name over 16 accepted")
	}
}
