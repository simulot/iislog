package iislogs

import (
	"time"

	"strings"

	"github.com/simulot/golib/file/walker"
	"github.com/simulot/golib/pipeline"
	"github.com/simulot/iislog/iis"
)

// ParserOperator create a operator for the pipeline in charge of
// opening the log parser and output matched lines
func (a *Application) ParserOperator() pipeline.Operator {
	filter := a.MakeLogRecordFilter()

	return func(in, out chan interface{}) {
		for i := range in {
			if item, ok := i.(walker.WalkItem); ok {
				if r, err := item.Reader(); err == nil {
					p := iis.NewLogParser(r)
					recChan := p.Parse(filter)
					for rec := range recChan {
						out <- rec
					}
				}
				item.Close()
			} else {
				panic("Expecting walker.WalkItem in pipeline.Operator filter")
			}
		}
	}
}

type filter struct {
	a *Application
}

func (f *filter) CheckFullLine(line *string) bool { return true }

func (f *filter) CheckDate(date time.Time) bool {
	return f.a.dateFrom.Before(date) && f.a.dateTo.After(date)
}
func (f *filter) CheckField(field string, value interface{}) (ret bool) {
	ret = true
	if ret && f.a.longQueries != 0 && field == "time-taken" {
		if value.(time.Duration) < f.a.longQueries {
			ret = false
		}
	}
	if ret && field == "cs-uri-stem" {
		if f.a.hideAssets {
			url := value.(string)
			if dot := strings.LastIndex(url, "."); dot >= 0 {
				switch strings.ToLower(url[dot:]) {
				case ".jpg", ".gif", ".png", ".css", ".js", ".ico", ".html":
					ret = false
				}
			}
		}
		if ret && len(f.a.urls) > 0 {
			ret = false
			for _, u := range f.a.urls {
				if strings.Contains(value.(string), u) {
					ret = true
					break
				}
			}
		}
	}
	if ret && len(f.a.users) > 0 && field == "cs-username" {
		ret = false
		for _, u := range f.a.users {
			if strings.Contains(value.(string), u) {
				ret = true
				break
			}
		}
	}
	return
}

func (f *filter) CheckError(status, substatus int) bool {
	if f.a.protocolError && status != 0 && (status < 400 || status >= 600 || (status == 401 && substatus == 2)) {
		return false
	}
	return true
}

func (a *Application) MakeLogRecordFilter() iis.RecordFilter {
	return &filter{a}
}
