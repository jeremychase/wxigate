package calcrain

import (
	"errors"
	"time"
)

type record struct {
	Time   time.Time
	Amount float64
}

type Data struct {
	Rain    []record
	PrevMax record
}

// Append amount to in-memory dataset.
func (d *Data) Append(amount float64, t time.Time) {
	r := record{
		Amount: amount,
		Time:   t,
	}

	// Keep track of when the day changes and update PrevMax.
	if len(d.Rain) > 1 {
		if d.Rain[len(d.Rain)-1].Time.Local().Weekday() != r.Time.Local().Weekday() {
			d.PrevMax = d.Rain[len(d.Rain)-1]
		}
	}

	d.Rain = append(d.Rain, r)
}

// RainLast24Hours returns rainfall over the trailing 24 hours based on
// recorded data. This prunes old data.
func (d *Data) RainLast24Hours(amount float64, t time.Time, threshold uint) (float64, error) {
	prev, err := d.prevNow(t, threshold)
	if err != nil {
		return 0.0, err
	}

	return d.PrevMax.Amount - prev + amount, err
}

// prevNow returns cumulative rain for day  previous to `t`. `threshold`
// represents the maximum age allowed.
func (d *Data) prevNow(t time.Time, threshold uint) (float64, error) {
	old := 0
	new := len(d.Rain) - 1

	for old <= new {
		mid := old + (new-old)/2

		if t.Sub(d.Rain[mid].Time) < time.Hour*24 {
			new = mid - 1
			if new == -1 {
				return 0.0, errors.New("insufficient data")
			}
		} else {
			old = mid + 1
		}
	}

	delta := t.Sub(d.Rain[new].Time)
	amount := d.Rain[new].Amount

	// prune old data
	if new > len(d.Rain)/4 {
		d.Rain = append([]record(nil), d.Rain[new/2:]...) // keep some for future calculations
	}

	if delta > time.Hour*24+time.Minute*time.Duration(threshold) {
		return 0.0, errors.New("prev too old")
	}

	return amount, nil
}
