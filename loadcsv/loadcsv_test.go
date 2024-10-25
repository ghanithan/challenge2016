package loadcsv

import "testing"

// Test using the comman `go test ./... -v`
func TestLoadCsv(t *testing.T) {
	t.Run("Test loading and parsing CSV", func(t *testing.T) {
		csv, err := LoadCsv("../cities.csv")
		if err != nil {
			t.Errorf("error occured in testing: %s", err)
		}
		t.Log(csv)
	})
}
