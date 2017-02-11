package iislogs

import (
	"fmt"

	"sync"

	"sort"

	"time"

	"github.com/simulot/golib/pipeline"
	"github.com/simulot/iislog/iis"
)

// DeduplicateOperator creates an operator that remove duplicates
// from flow, and sort it by date for application's pipeline
func (a *Application) DeduplicateOperator() pipeline.Operator {
	lock := sync.RWMutex{}
	seen := map[string]bool{}
	records := iis.LogRecords{}

	return func(in, out chan interface{}) {
		for i := range in {
			if item, ok := i.(*iis.LogRecord); ok {
				lock.RLock()
				_, dejaVu := seen[item.Raw]
				lock.RUnlock()
				if !dejaVu {
					lock.Lock()
					seen[item.Raw] = true
					records = append(records, item)
					lock.Unlock()
				}
			} else {
				panic("Expecting *iis.LogRecord in pipeline.Operator OutputOperator")
			}
		}

		// At this point, we have selected all items
		sort.Sort(records)
		for _, item := range records {
			out <- item
		}
	}
}

// OutputOperator creates an output for application's pipeline
func (a *Application) OutputOperator() pipeline.Operator {
	return func(in, out chan interface{}) {
		fmt.Print("date;")
		fmt.Print("status;")
		fmt.Print("s-ip;")
		fmt.Print("cs-username;")
		fmt.Print("cs-uri-stem;")
		fmt.Print("cs-uri-query;")
		fmt.Print("time-taken(ms);")
		fmt.Print("time-taken;")
		fmt.Print("status-label")
		fmt.Print("\r\n")

		for i := range in {
			if item, ok := i.(*iis.LogRecord); ok {
				fmt.Print(item.DateTime.Format(`"2006-01-02 15:04:05"`), ";")
				fmt.Print(item.Get("status"), ";")
				fmt.Print(item.Server, ";")
				fmt.Print(item.User, ";")
				fmt.Print(item.URI, ";")
				fmt.Print(`"`, item.Query, `"`, ";")
				fmt.Print(int(item.Get("time-taken").(time.Duration)/time.Millisecond), ";")
				fmt.Print(item.Get("time-taken"), ";")
				fmt.Print(`"`, item.Get("status-label"), `"`)
				fmt.Print("\r\n")
			} else {
				panic("Expecting *iis.LogRecord in pipeline.Operator OutputOperator")
			}
		}
	}
}
