package web

import (
	"bytes"
	"encoding/csv"
	"strings"
	"testing"
)

func TestWriteNomadeezeeCSV(t *testing.T) {
	source := strings.NewReader("title,descriptions,category,latitude,longitude,address,complete_address\n" +
		"Coffee Place,Great coffee,Coffee shop,16.0544,108.2022,1 Main St,\"{\"\"city\"\":\"\"Da Nang\"\",\"\"country\"\":\"\"Vietnam\"\"}\"\n")

	var destination bytes.Buffer
	if err := writeNomadeezeeCSV(&destination, source); err != nil {
		t.Fatalf("writeNomadeezeeCSV: %v", err)
	}

	rows, err := csv.NewReader(&destination).ReadAll()
	if err != nil {
		t.Fatalf("read output: %v", err)
	}

	if len(rows) != 2 {
		t.Fatalf("expected header and one row, got %d", len(rows))
	}

	want := []string{"Coffee Place", "Great coffee", "cafe", "16.0544", "108.2022", "1 Main St", "Vietnam", "Da Nang", "true", "false"}
	if got := rows[1]; !equalStringSlices(got, want) {
		t.Fatalf("unexpected output row:\n got: %#v\nwant: %#v", got, want)
	}
}

func TestNomadeezeeCategoryFallsBackToOther(t *testing.T) {
	if got := nomadeezeeCategory("Pet groomer"); got != "other" {
		t.Fatalf("expected other, got %q", got)
	}
}

func equalStringSlices(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}

	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}

	return true
}
