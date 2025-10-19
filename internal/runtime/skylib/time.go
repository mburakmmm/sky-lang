package skylib

import (
	"time"
)

// Now returns current Unix timestamp in milliseconds
func TimeNow() int64 {
	return time.Now().UnixMilli()
}

// NowNano returns current Unix timestamp in nanoseconds
func TimeNowNano() int64 {
	return time.Now().UnixNano()
}

// Sleep sleeps for specified milliseconds
func TimeSleep(ms int) {
	time.Sleep(time.Duration(ms) * time.Millisecond)
}

// Since returns elapsed time in milliseconds since timestamp
func TimeSince(timestamp int64) int64 {
	return time.Now().UnixMilli() - timestamp
}

// Format formats a timestamp
func TimeFormat(timestamp int64, layout string) string {
	t := time.UnixMilli(timestamp)
	
	// Common layouts
	switch layout {
	case "ISO8601":
		return t.Format(time.RFC3339)
	case "RFC3339":
		return t.Format(time.RFC3339)
	case "RFC822":
		return t.Format(time.RFC822)
	case "ANSIC":
		return t.Format(time.ANSIC)
	case "UnixDate":
		return t.Format(time.UnixDate)
	case "RubyDate":
		return t.Format(time.RubyDate)
	default:
		// Custom format
		return t.Format(layout)
	}
}

// Parse parses a time string
func TimeParse(layout, value string) (int64, error) {
	var t time.Time
	var err error
	
	switch layout {
	case "ISO8601", "RFC3339":
		t, err = time.Parse(time.RFC3339, value)
	case "RFC822":
		t, err = time.Parse(time.RFC822, value)
	case "ANSIC":
		t, err = time.Parse(time.ANSIC, value)
	case "UnixDate":
		t, err = time.Parse(time.UnixDate, value)
	default:
		t, err = time.Parse(layout, value)
	}
	
	if err != nil {
		return 0, err
	}
	
	return t.UnixMilli(), nil
}

// Measure measures execution time of a function
func TimeMeasure(fn func()) int64 {
	start := time.Now()
	fn()
	return time.Since(start).Milliseconds()
}

// AddDuration adds duration to timestamp
func TimeAdd(timestamp int64, durationMs int64) int64 {
	return timestamp + durationMs
}

// Diff calculates difference between two timestamps
func TimeDiff(t1, t2 int64) int64 {
	return t1 - t2
}

