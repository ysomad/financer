package date

import (
	"fmt"
	"time"
)

func Parse(s string) (time.Time, error) {
	var (
		fmts = [3]string{"02.01.2006", "02.01.06", "02.01"}
		date time.Time
		err  error
	)

	for _, layout := range fmts {
		date, err = time.Parse(layout, s)
		if err == nil {
			if date.Year() == 0 {
				date = time.Date(time.Now().Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.Local)
			}

			return date, nil
		}
	}

	return date, fmt.Errorf("invalid date format: %s", s)
}
