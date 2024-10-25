package config

import (
	"reflect"
	"testing"

	"github.com/ghanithan/challenge2016/instrumentation"
)

func TestGetConfig(t *testing.T) {
	t.Run(
		"Testing getConfig",
		func(t *testing.T) {
			want := &Config{
				Data: Data{
					FilePath: "../cities.csv",
				},
			}

			logger := instrumentation.InitInstruments()

			got, err := GetConfig(logger)
			if err != nil {
				t.Fatalf("Error in fetching config: %s", err)
			}
			if !reflect.DeepEqual(want, got) {
				t.Fatalf("expected %q got %q", want, got)

			}
		},
	)
}
