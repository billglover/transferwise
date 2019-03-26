package transferwise

import (
	"context"
	"net/http"
	"time"
)

// Rate is an exchange rate between two currencies at a point in time.
type Rate struct {
	Rate   float64
	Source string
	Target string
	Time   Time
}

// RatesFilter contains a set of options that can be passed to a Rates
// query to filter the result set returned.
type RatesFilter struct {
	Source string
	Target string
	Time   time.Time
	From   time.Time
	To     time.Time
	Group  string
}

// Rates returns the latest exchange rates of all currencies.
func (c *Client) Rates(ctx context.Context, f *RatesFilter) ([]Rate, *http.Response, error) {
	req, err := c.NewRequest("GET", "/v1/rates", nil)
	if err != nil {
		return nil, nil, err
	}

	if f != nil {
		q := req.URL.Query()

		if f.Source != "" {
			q.Set("source", f.Source)
		}

		if f.Target != "" {
			q.Set("target", f.Target)
		}

		req.URL.RawQuery = q.Encode()
	}

	var rates []Rate
	resp, err := c.Do(ctx, req, &rates)
	if err != nil {
		return nil, resp, err
	}
	return rates, resp, nil
}
