package iislog

import (
	"time"

	"github.com/simulot/golib/file/walker"
	"github.com/simulot/golib/pipeline"
)

// FileFilterOperator create a file filter for the application pipeline
func (a *Application) FileFilterOperator() pipeline.Operator {
	return func(in, out chan interface{}) {
		from, to := a.dateFrom.Truncate(24*time.Hour), a.dateTo.Truncate(24*time.Hour).AddDate(0, 0, 1)
		for i := range in {
			if item, ok := i.(walker.WalkItem); ok {
				if fd, err := time.ParseInLocation("u_ex060102.log", item.Name(), time.UTC); err == nil {
					// Check if the file date fits with searched date range
					if from.Before(fd) && fd.Before(to) {
						out <- item
					}
				} else {
					item.Close()
				}
			} else {
				panic("Expecting walker.WalkItem in pipeline.Operator filter")
			}
		}
	}
}
