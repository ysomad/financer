package money

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type M struct {
	Units int64
	Nanos int32
}

func ParseString(s string) (M, error) {
	parts := strings.Split(s, ".")

	if len(parts) == 0 || len(parts) > 2 {
		return M{}, errors.New("invalid money format")
	}

	var (
		money M
		err   error
	)

	money.Units, err = strconv.ParseInt(parts[0], 10, 32)
	if err != nil {
		return M{}, fmt.Errorf("units not parsed: %w", err)
	}

	// money got nanos
	if len(parts) == 2 {
		sb := &strings.Builder{}
		sb.Grow(9)
		sb.WriteString(parts[1])

		for sb.Len() < 9 {
			sb.WriteString("0")
		}

		nanosStr := sb.String()

		nanos64, err := strconv.ParseInt(nanosStr[:9], 10, 32)
		if err != nil {
			return M{}, fmt.Errorf("nanos not parsed: %w", err)
		}

		if money.Units > 0 {
			money.Nanos = int32(nanos64)
		} else {
			money.Nanos = -int32(nanos64)
		}
	}

	return money, nil
}
