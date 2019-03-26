package transferwise

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"testing"
	"time"
)

var t1, _ = time.Parse("2006-01-02T15:04:05Z0700", "2018-08-31T10:43:31+0000")
var testCases = []struct {
	query RatesFilter
	rates []Rate
}{
	{
		query: RatesFilter{Source: "EUR", Target: "USD"},
		rates: []Rate{
			Rate{Source: "EUR", Target: "USD", Rate: 1.166, Time: Time{t1}},
		},
	},
}

func TestRates(t *testing.T) {
	for _, tc := range testCases {
		tw, mux, _, teardown := setup()
		defer teardown()

		mux.HandleFunc("/v1/rates", func(w http.ResponseWriter, r *http.Request) {
			if got, want := r.Method, http.MethodGet; got != want {
				t.Errorf("unexpected method: got %s, want %s", got, want)
			}

			if got, want := r.URL.Query().Get("source"), tc.query.Source; got != want {
				t.Errorf("unexpected source: got %s, want %s", got, want)
			}

			if got, want := r.URL.Query().Get("target"), tc.query.Target; got != want {
				t.Errorf("unexpected target: got %s, want %s", got, want)
			}

			fmt.Fprintln(w, `[
				{
					"rate": 1.166,
					"source": "EUR",
					"target": "USD",
					"time": "2018-08-31T10:43:31+0000"
				}
			]`)
		})

		got, _, err := tw.Rates(context.Background(), &tc.query)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if reflect.DeepEqual(got, tc.rates) == false {
			t.Errorf("unexpected response:\n\tgot:  %+v\n\twant: %+v", got, tc.rates)
		}
	}
}
