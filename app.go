package iislog

import (
	"time"

	_ "github.com/simulot/golib/file/walker/zipwalker" //register zip walker
	"github.com/simulot/golib/pipeline"
)

// Application represents the application state and its parameters
type Application struct {
	files            []string      // Files, paths, archives to be explored
	dateFrom, dateTo time.Time     // Set exploration time limits
	protocolError    bool          // indicates to filter 4xx and 5xx errors
	hideAssets       bool          // indicate to filer gif, png, css, js
	longQueries      time.Duration // search queries longer than this
	urls             []string      // URL to be reported. Cumulative.
	users            []string      // List of user concerned. Cumulative
}

// Run runs the application
func (a *Application) Run() {
	in := make(chan interface{})
	go func() {
		for _, file := range a.files {
			in <- file // Injects CLI arguments into the pipeline
		}
		close(in) // We are done
	}()

	pipe := pipeline.NewFlow(
		// Expands arguments having wild cards into flow of path
		pipeline.GlobOperator(),

		// Takes a path (file, folder, archive) and makes a walker on it
		pipeline.FolderToWalkersOperator(),

		pipeline.NewParallelFlow(
			8,

			// Walks through the walker and makes a flow of items (folder files, archive items)
			pipeline.WalkOperator(),

			// Excludes everything but IIS log files
			pipeline.FileMaskOperator("u_ex*.log"),

			// Excludes files which date is outside time frame
			a.FileFilterOperator(),

			// Parses log file and emits a flow of log records
			a.ParserOperator(),
		),

		// Remove multiples occurrences of an entry
		a.DeduplicateOperator(),

		// Outputs log records
		a.OutputOperator(),
	)
	<-pipe.Run(in)
}
