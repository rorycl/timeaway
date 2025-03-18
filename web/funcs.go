package web

import "time"

// funcs provide some commonly used template functions

// yearsAgo provides a function for adding or removing years from the
// provided date.
func yearsAgo(d time.Time, years int) time.Time {
	return time.Date(d.Year()+years, d.Month(), d.Day(), d.Hour(), d.Minute(), d.Second(), d.Nanosecond(), d.Location())
}

// dateStr provides the standard date string format for this web app.
func dateStr(d time.Time) string {
	return d.Format(time.DateOnly)
}

// webFuncMap provides a map suitable for providing to template.Funcs.
var webFuncMap map[string]any = map[string]any{
	"yearsAgo": yearsAgo,
	"dateStr":  dateStr,
}
