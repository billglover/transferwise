package transferwise

import "time"

// Time is a custom type used to handle time formats used by the
// Transferwise API.
type Time struct {
	time.Time
}

// UnmarshalJSON unmarshalls the time format used by the TransferWise API.
// It first attempts to use RFC3339 and if that fails attempts to use a
// similar format without the colon in the timezone offset.
// Source: https://stackoverflow.com/questions/39179910/unmarshal-incorrectly-formated-datetime-in-golang
func (t *Time) UnmarshalJSON(b []byte) (err error) {
	s := string(b)
	s = s[1 : len(s)-1]

	p, err := time.Parse(time.RFC3339, s)
	if err != nil {
		p, err = time.Parse("2006-01-02T15:04:05Z0700", s)
	}
	t.Time = p
	return
}
