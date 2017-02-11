package iislog

import (
	"time"

	"os"

	"gopkg.in/alecthomas/kingpin.v2"
)

// ParseCommandLine manage command line parameters and arguments
func (a *Application) ParseCommandLine() (string, error) {
	// some default value
	zero := time.Now().UTC().Truncate(24 * time.Hour)
	a.dateFrom = time.Date(1900, 01, 01, 00, 0, 0, 0, time.UTC)
	a.dateTo = time.Date(9999, 12, 31, 23, 59, 59, 999999999, time.UTC)

	app := kingpin.New("iislog", "a tool for searching in IIS logs files.\n\tAuthor jfc@responsiveconsulting.fr\n")
	setTimeValue(newTimeValue(&a.dateFrom), app.Flag("from", "get logs from 'DATETIME' UTC").PlaceHolder("DATETIME"))
	setTimeValue(newTimeValue(&a.dateTo), app.Flag("to", "get logs to 'DATETIME' UTC").PlaceHolder("DATETIME"))

	fromDaysAgo, toDaysAgo := 0, 0
	app.Flag("from-days-ago", "get logs from DAYS ago").PlaceHolder("DAYS").Action(func(c *kingpin.ParseContext) error {
		a.dateFrom = zero.Add(time.Duration(fromDaysAgo) * time.Hour * -24)
		return nil
	}).IntVar(&fromDaysAgo)
	app.Flag("to-days-ago", "get logs until DAYS before today").PlaceHolder("DAYS").Action(func(c *kingpin.ParseContext) error {
		a.dateTo = zero.Add(time.Duration(toDaysAgo) * time.Hour * -24)
		return nil
	}).IntVar(&toDaysAgo)

	since := time.Duration(0)
	app.Flag("since", "get logs since DURATION. DURATION can be like 2s, 24h...").PlaceHolder("DURATION").Action(func(c *kingpin.ParseContext) error {
		a.dateFrom = time.Now().UTC().Add(time.Duration(since))
		return nil
	}).DurationVar(&since)

	app.Flag("url", "Reports lines containing url. Several --url options can be given. Lines are reported whenever one url matches").
		StringsVar(&a.urls)
	app.Flag("user", "Reports lines from authenticated USER. Several --user options can be given. Lines are reported whenever an user matches").
		StringsVar(&a.users)

	// app.Flag("search", "search string in log").Short('s').StringVar(&a.searchedString)

	app.Flag("errors", "filter logs on protocol errors (4xx and 5xx)").BoolVar(&a.protocolError)

	app.Flag("hide-assets", "hide assets (html,gif,ico,css,jpg,png,js) from result list").
		BoolVar(&a.hideAssets)

	app.Flag("long-queries", "show queries longer than 'DURATION'. Accepted values like 200ms, 3s, 1m...").PlaceHolder("DURATION").
		DurationVar(&a.longQueries)

	app.Arg("file", "file, path, zip archive").Required().StringsVar(&a.files)

	cmd, err := app.Parse(os.Args[1:])

	return cmd, err

}

// -------- time value -------------------------------------------
type timeValue struct {
	v *time.Time
}

const timeValueFormat = "2006-01-02 15:04:05"

func setTimeValue(v *timeValue, s kingpin.Settings) {
	s.SetValue(v)
}

func newTimeValue(p *time.Time) *timeValue {
	return &timeValue{p}
}

func (f *timeValue) Set(s string) error {
	x, err := time.ParseInLocation(timeValueFormat, s, time.UTC)
	if err != nil {
		return err
	}
	*(f.v) = x
	return nil
}
func (f *timeValue) String() string { return f.v.Format(timeValueFormat) }
