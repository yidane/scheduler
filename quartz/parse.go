package quartz

import (
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
	"time"
)

// Parse returns a new crontab schedule representing the given spec.
// It panics with a descriptive error if the spec is not valid.
//
// It accepts
//   - Full crontab specs, e.g. "* * * * * ?"
//   - Descriptors, e.g. "@midnight", "@every 1h30m"
func Parse(spec string) (Schedule, error) {
	if spec[0] == '@' {
		return parseDescriptor(spec)
	}

	// Split on whitespace.  We require 5 or 6 fields.
	// (second) (minute) (hour) (day of month) (month) (day of week, optional)
	fields := strings.Fields(spec)
	if len(fields) != 5 && len(fields) != 6 {
		log.Printf("Expected 5 or 6 fields, found %d: %s\n", len(fields), spec)
		return nil, fmt.Errorf("Expected 5 or 6 fields, found %d: %s", len(fields), spec)
	}

	// If a sixth field is not provided (DayOfWeek), then it is equivalent to star.
	if len(fields) == 5 {
		fields = append(fields, "*")
	}
	sec, err := getField(fields[0], seconds)
	if err != nil {
		return nil, err
	}
	min, err := getField(fields[1], minutes)
	if err != nil {
		return nil, err
	}

	hou, err := getField(fields[2], hours)
	if err != nil {
		return nil, err
	}
	doom, err := getField(fields[3], dom)
	if err != nil {
		return nil, err
	}
	mon, err := getField(fields[4], months)
	if err != nil {
		return nil, err
	}
	doow, err := getField(fields[5], dow)
	if err != nil {
		return nil, err
	}

	schedule := &SpecSchedule{
		Second: sec,
		Minute: min,
		Hour:   hou,
		Dom:    doom,
		Month:  mon,
		Dow:    doow,
	}

	return schedule, nil
}

func ValidSpec(spec string) bool {
	return true
}

// getField returns an Int with the bits set representing all of the times that
// the field represents.  A "field" is a comma-separated list of "ranges".
func getField(field string, r bounds) (uint64, error) {
	// list = range {"," range}
	var bits uint64
	ranges := strings.FieldsFunc(field, func(r rune) bool {
		return r == ','
	})
	for _, expr := range ranges {
		bit, err := getRange(expr, r)
		if err != nil {
			return 0, err
		}
		bits |= bit
	}
	return bits, nil
}

// getRange returns the bits indicated by the given expression:
//   number | number "-" number [ "/" number ]
func getRange(expr string, r bounds) (uint64, error) {

	var (
		start, end, step uint
		rangeAndStep     = strings.Split(expr, "/")
		lowAndHigh       = strings.Split(rangeAndStep[0], "-")
		singleDigit      = len(lowAndHigh) == 1
	)

	var extra_star uint64
	if lowAndHigh[0] == "*" || lowAndHigh[0] == "?" {
		start = r.min
		end = r.max
		extra_star = starBit
	} else {
		start, err := parseIntOrName(lowAndHigh[0], r.names)
		if err != nil {
			return 0, err
		}
		switch len(lowAndHigh) {
		case 1:
			end = start
		case 2:
			var err error
			end, err = parseIntOrName(lowAndHigh[1], r.names)
			if err != nil {
				return 0, err
			}
		default:
			log.Println("Too many hyphens: %s", expr)
			return 0, fmt.Errorf("Too many hyphens: %s", expr)
		}
	}

	switch len(rangeAndStep) {
	case 1:
		step = 1
	case 2:
		var err error
		step, err = mustParseInt(rangeAndStep[1])
		if err != nil {
			return 0, err
		}

		// Special handling: "N/step" means "N-max/step".
		if singleDigit {
			end = r.max
		}
	default:
		log.Panicf("Too many slashes: %s", expr)
		return 0, fmt.Errorf("Too many slashes: %s", expr)
	}

	if start < r.min {
		log.Println("Beginning of range (%d) below minimum (%d): %s", start, r.min, expr)
		return 0, fmt.Errorf("Beginning of range (%d) below minimum (%d): %s", start, r.min, expr)
	}
	if end > r.max {
		log.Println("End of range (%d) above maximum (%d): %s", end, r.max, expr)
		return 0, fmt.Errorf("End of range (%d) above maximum (%d): %s", end, r.max, expr)
	}
	if start > end {
		log.Println("Beginning of range (%d) beyond end of range (%d): %s", start, end, expr)
		return 0, fmt.Errorf("Beginning of range (%d) beyond end of range (%d): %s", start, end, expr)
	}

	return getBits(start, end, step) | extra_star, nil
}

// parseIntOrName returns the (possibly-named) integer contained in expr.
func parseIntOrName(expr string, names map[string]uint) (uint, error) {
	if names != nil {
		if namedInt, ok := names[strings.ToLower(expr)]; ok {
			return namedInt, nil
		}
	}
	return mustParseInt(expr)
}

// mustParseInt parses the given expression as an int or panics.
func mustParseInt(expr string) (uint, error) {
	num, err := strconv.Atoi(expr)
	if err != nil {
		log.Println("Failed to parse int from %s: %s", expr, err)
		return 0, fmt.Errorf("Failed to parse int from %s: %s", expr, err)
	}
	if num < 0 {
		log.Println("Negative number (%d) not allowed: %s", num, expr)
		return 0, fmt.Errorf("Negative number (%d) not allowed: %s", num, expr)
	}

	return uint(num), nil
}

// getBits sets all bits in the range [min, max], modulo the given step size.
func getBits(min, max, step uint) uint64 {
	var bits uint64

	// If step is 1, use shifts.
	if step == 1 {
		return ^(math.MaxUint64 << (max + 1)) & (math.MaxUint64 << min)
	}

	// Else, use a simple loop.
	for i := min; i <= max; i += step {
		bits |= 1 << i
	}
	return bits
}

// all returns all bits within the given bounds.  (plus the star bit)
func all(r bounds) uint64 {
	return getBits(r.min, r.max, 1) | starBit
}

// parseDescriptor returns a pre-defined schedule for the expression, or panics
// if none matches.
func parseDescriptor(spec string) (Schedule, error) {
	switch spec {
	case "@yearly", "@annually":
		return &SpecSchedule{
			Second: 1 << seconds.min,
			Minute: 1 << minutes.min,
			Hour:   1 << hours.min,
			Dom:    1 << dom.min,
			Month:  1 << months.min,
			Dow:    all(dow),
		}, nil

	case "@monthly":
		return &SpecSchedule{
			Second: 1 << seconds.min,
			Minute: 1 << minutes.min,
			Hour:   1 << hours.min,
			Dom:    1 << dom.min,
			Month:  all(months),
			Dow:    all(dow),
		}, nil

	case "@weekly":
		return &SpecSchedule{
			Second: 1 << seconds.min,
			Minute: 1 << minutes.min,
			Hour:   1 << hours.min,
			Dom:    all(dom),
			Month:  all(months),
			Dow:    1 << dow.min,
		}, nil
	case "@daily", "@midnight":
		return &SpecSchedule{
			Second: 1 << seconds.min,
			Minute: 1 << minutes.min,
			Hour:   1 << hours.min,
			Dom:    all(dom),
			Month:  all(months),
			Dow:    all(dow),
		}, nil
	case "@hourly":
		return &SpecSchedule{
			Second: 1 << seconds.min,
			Minute: 1 << minutes.min,
			Hour:   all(hours),
			Dom:    all(dom),
			Month:  all(months),
			Dow:    all(dow),
		}, nil
	}
	const every = "@every"
	if strings.HasPrefix(spec, every) {
		duration, err := time.ParseDuration(spec[len(every):])
		if err != nil {
			log.Printf("Failed to parse duration %s: %s\n", spec, err)
			return nil, err
		}
		return Every(duration), nil
	}

	log.Printf("Unrecognized descriptor: %s\n", spec)
	return nil, fmt.Errorf("Unrecognized descriptor: %s", spec)
}
