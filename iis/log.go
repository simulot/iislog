package iis

import (
	"bufio"
	"io"
	"strconv"
	"strings"
	"time"
)

// LogParser is a parser for IISlogs
type LogParser struct {
	r      *bufio.Reader
	fields []string
}

// LogRecord is an individual log line
type LogRecord struct {
	Raw        string            // Raw log line
	DateTime   time.Time         // Log event time UTC
	User       string            // Identified user
	Server     string            // Server's IP
	Site       string            // Begining of the url
	URI        string            // URI
	Query      string            // Request parameters
	Status     int               // Main status code
	SubStatus  int               // Sub status code
	TimeTaken  time.Duration     // Request time taken
	Other      map[string]string // Other extracted fields
	filter     RecordFilter      // Inject filter logic
	date, hour string            // temporary storage for reading date and time from separate fields
}

// NewLogParser creates an instance of IIS LogParser
func NewLogParser(r io.Reader) *LogParser {
	return &LogParser{
		r: bufio.NewReader(r),
	}
}

const (
	fieldsPrefix = "#Fields: "
	datePrefix   = "#Date: "
	tsFormat     = "2006-01-02 15:04:05"
)

// Parse the log and emits log records on the out chan
func (l *LogParser) Parse(filter RecordFilter) chan *LogRecord {
	out := make(chan *LogRecord)
	go l.doParse(out, filter)
	return out
}

// doParse does the actual work of parsing the log file
func (l *LogParser) doParse(out chan *LogRecord, filter RecordFilter) {
	logDate := time.Time{}
	s, err := l.r.ReadString('\n')

	for err == nil || (err == io.EOF && len(s) > 0) {
		for i := len(s) - 1; i > 0 && (s[i] == '\r' || s[i] == '\n'); i-- {
			s = s[:i]
		}
		switch {

		// #Date: get log date
		case strings.HasPrefix(s, datePrefix):
			s = s[len(datePrefix):]
			logDate, err = time.ParseInLocation(tsFormat, s, time.UTC)

		// #Fields : get fields definition, may have changed
		case strings.HasPrefix(s, fieldsPrefix):
			s = s[len(fieldsPrefix):]
			l.fields = []string{}
			mark := 0
			for i := 0; i < len(s); i++ {
				if i == len(s)-1 || s[i] == ' ' {
					if i == len(s)-1 {
						i++ // I don't understand wy it's necessary!
					}
					l.fields = append(l.fields, s[mark:i])
					mark = i + 1
				}
			}

		// skip empty lines or other commented out lines
		case len(s) < 2, s[0] == '#':
			{
			}
		default:
			{
				// Do parsing work only if the log date is in time frame and if
				// a full text pattern is recognized
				if filter.CheckDate(logDate) && filter.CheckFullLine(&s) {
					r := NewLogRecord(s, filter)
					fieldIndex := 0
					mark := 0
					selected := true
					for i := 0; i < len(s); i++ {
						if i == len(s)-1 || s[i] == ' ' {
							if i == len(s)-1 {
								i++ // don't get this!
							}
							field := s[mark:i]
							mark = i + 1
							if !r.Set(l.fields[fieldIndex], field) {
								// the record parsing is abandonned
								// as soon as a field is rejected by the filter
								selected = false
								break
							}
							fieldIndex++
						}
					}
					if selected {
						out <- r
					}
				}
			}
		}
		s, err = l.r.ReadString('\n')
	}
	close(out)
}

// NewLogRecord creates a new instance of log record
func NewLogRecord(raw string, filter RecordFilter) *LogRecord {
	if filter == nil {
		// Provides a nop filter when given is nil
		filter = &NoFilter{}
	}
	return &LogRecord{
		Raw:    raw,
		filter: filter,
		Other:  make(map[string]string),
	}
}

// Set record field value, and do some additional logic
// return true if the record is still selected
func (r *LogRecord) Set(field, value string) bool {
	var err error
	var i int

	switch field {
	case "DateTime":
		r.DateTime, err = time.ParseInLocation(tsFormat, r.date+" "+r.hour, time.UTC)
		return r.filter.CheckDate(r.DateTime)
	case "date":
		r.date = value
		if len(r.hour) > 0 {
			return r.Set("DateTime", "")
		}
		return true
	case "time":
		r.hour = value
		if len(r.date) > 0 {
			return r.Set("DateTime", "")
		}
		return true
	case "s-ip":
		r.Server = value
		return r.filter.CheckField(field, value)
	case "cs-username":
		r.User = value
		return r.filter.CheckField(field, value)
	case "cs-uri-stem":
		r.URI = value
		if i := strings.Index(value[1:], "/"); i >= 0 {
			r.Site = value[0 : i+1]
		}
		return r.filter.CheckField(field, value)
	case "cs-uri-query":
		r.Query = value
		return r.filter.CheckField(field, value)
	case "sc-status":
		r.Status, err = strconv.Atoi(value)
		return r.filter.CheckError(r.Status, r.SubStatus)
	case "sc-substatus":
		r.SubStatus, err = strconv.Atoi(value)
		return r.filter.CheckError(r.Status, r.SubStatus)
	case "time-taken":
		i, err = strconv.Atoi(value)
		if err == nil {
			r.TimeTaken = time.Millisecond * time.Duration(i)
		}
		return r.filter.CheckField(field, r.TimeTaken)

	default:
		r.Other[field] = value
		return r.filter.CheckField(field, value)
	}
}

// Get log record value per field name
func (r *LogRecord) Get(field string) interface{} {
	switch field {
	case "DateTime":
		return r.DateTime
	case "date":
		return r.DateTime.Format("2006-01-02")
	case "time":
		return r.DateTime.Format("15:04:05")
	case "s-ip":
		return r.Server
	case "cs-username":
		return r.User
	case "cs-uri-stem":
		return r.URI
	case "cs-uri-query":
		return r.Query
	case "sc-status":
		return r.Status
	case "sc-substatus":
		return r.SubStatus
	case "time-taken":
		return r.TimeTaken //int(r.TimeTaken / time.Microsecond)
	case "status":
		return strconv.Itoa(r.Status) + "." + strconv.Itoa(r.SubStatus)
	case "status-label":
		return iisStatus(r.Status, r.SubStatus)
	case "site":
		return r.Site
	default:
		return r.Other[field]
	}
}

// IsAnError returns true when status denotes a protocol error
func (r *LogRecord) IsAnError() bool {
	return r.Status >= 400 && r.Status < 600
}

// Details on status at https://support.microsoft.com/en-us/help/943891/the-http-status-code-in-iis-7.0,-iis-7.5,-and-iis-8.0
var protocolExtendedError = map[int]map[int]string{
	100: {0: "Continue"},
	101: {0: "Switching protocols"},
	200: {0: "OK. The client request has succeeded"},
	201: {0: "Created"},
	202: {0: "Accepted"},
	203: {0: "Nonauthoritative information"},
	204: {0: "No content"},
	205: {0: "Reset content"},
	206: {0: "Partial content"},
	301: {0: "Moved permanently"},
	302: {0: "Object moved"},
	304: {0: "Not modified"},
	307: {0: "Temporary redirect"},
	400: {
		0: "Bad request",
		1: "Invalid Destination Header",
		2: "Invalid Depth Header",
		3: "Invalid If Header",
		4: "Invalid Overwrite Header",
		5: "Invalid Translate Header",
		6: "Invalid Request Body",
		7: "Invalid Content Length",
		8: "Invalid Timeout",
		9: "Invalid Lock Token",
	},
	401: {
		0: "Access denied",
		1: "Logon failed",
		2: "Logon failed due to server configuration",
		3: "Unauthorized due to ACL on resource",
		4: "Authorization failed by filter",
		5: "Authorization failed by ISAPI/CGI application",
	},
	403: {
		0:   "Forbidden",
		1:   "Execute access forbidden",
		2:   "Read access forbidden",
		3:   "Write access forbidden",
		4:   "SSL required",
		5:   "SSL 128 required",
		6:   "IP address rejected",
		7:   "Client certificate required",
		8:   "Site access denied",
		9:   "Forbidden: Too many clients are trying to connect to the web server",
		10:  "Forbidden: web server is configured to deny Execute access",
		11:  "Forbidden: Password has been changed",
		12:  "Mapper denied access",
		13:  "Client certificate revoked",
		14:  "Directory listing denied",
		15:  "Forbidden: Client access licenses have exceeded limits on the web server",
		16:  "Client certificate is untrusted or invalid",
		17:  "Client certificate has expired or is not yet valid",
		18:  "Cannot execute requested URL in the current application pool",
		19:  "Cannot execute CGI applications for the client in this application pool",
		20:  "Forbidden: Passport logon failed",
		21:  "Forbidden: Source access denied",
		22:  "Forbidden: Infinite depth is denied",
		502: "Forbidden: Too many requests from the same client IP; Dynamic IP Restriction limit reached",
	},
	404: {
		0:  "Not found",
		1:  "Site Not Found",
		2:  "ISAPI or CGI restriction",
		3:  "MIME type restriction",
		4:  "No handler configured",
		5:  "Denied by request filtering configuration",
		6:  "Verb denied",
		7:  "File extension denied",
		8:  "Hidden namespace",
		9:  "File attribute hidden",
		10: "Request header too long",
		11: "Request contains double escape sequence",
		12: "Request contains high-bit characters",
		13: "Content length too large",
		14: "Request URL too long",
		15: "Query string too long",
		16: "DAV request sent to the static file handler",
		17: "Dynamic content mapped to the static file handler via a wildcard MIME mapping",
		18: "Querystring sequence denied",
		19: "Denied by filtering rule",
		20: "Too Many URL Segments",
	},
	405: {0: "Method Not Allowed"},
	406: {0: "Client browser does not accept the MIME type of the requested page"},
	408: {0: "Request timed out"},
	412: {0: "Precondition failed"},
	500: {
		0:   "Module or ISAPI error occurred",
		11:  "Application is shutting down on the web server",
		12:  "Application is busy restarting on the web server",
		13:  "Web server is too busy",
		15:  "Direct requests for Global.asax are not allowed",
		19:  "Configuration data is invalid",
		21:  "Module not recognized",
		22:  "An ASP.NET httpModules configuration does not apply in Managed Pipeline mode",
		23:  "An ASP.NET httpHandlers configuration does not apply in Managed Pipeline mode",
		24:  "An ASP.NET impersonation configuration does not apply in Managed Pipeline mode",
		50:  "A rewrite error occurred during RQ_BEGIN_REQUEST notification handling. A configuration or inbound rule execution error occurred",
		51:  "A rewrite error occurred during GL_PRE_BEGIN_REQUEST notification handling. A global configuration or global rule execution error occurred",
		52:  "A rewrite error occurred during RQ_SEND_RESPONSE notification handling. An outbound rule execution occurred",
		53:  "A rewrite error occurred during RQ_RELEASE_REQUEST_STATE notification handling. An outbound rule execution error occurred. The rule is configured to be executed before the output user cache gets updated",
		100: "Internal ASP error",
	},
	501: {0: "Header values specify a configuration that is not implemented"},
	502: {
		0: "Web server received an invalid response while acting as a gateway or proxy",
		1: "CGI application timeout",
		2: "Bad gateway: Premature Exit",
		3: "Bad Gateway: Forwarder Connection Error (ARR)",
		4: "Bad Gateway: No Server (ARR)",
	},
	503: {
		0: "Application pool unavailable",
		2: "Concurrent request limit exceeded",
		3: "ASP.NET queue full",
	},
}

func iisStatus(status, substatus int) string {
	if m, ok := protocolExtendedError[status]; ok {
		if s, ok := m[substatus]; ok {
			return s
		}
		return m[0] + ", unknown substatus(" + strconv.Itoa(substatus) + ")"
	}
	return "Unknown status(" + strconv.Itoa(status) + "." + strconv.Itoa(substatus) + ")"
}

// RecordFilter is an interface for a filter, defined in an higher level, and injected
// in the log parser to discard records as soon we know it won't be emitted.
// For each function, returning true means keep the record, false means discard it.
type RecordFilter interface {
	CheckFullLine(line *string) bool
	CheckDate(date time.Time) bool
	CheckField(field string, value interface{}) bool
	CheckError(status, substatus int) bool
}

// NoFilter is a filter that filter nothing
type NoFilter struct{}

func (*NoFilter) CheckFullLine(line *string) bool                 { return true }
func (*NoFilter) CheckDate(date time.Time) bool                   { return true }
func (*NoFilter) CheckField(field string, value interface{}) bool { return true }
func (*NoFilter) CheckError(status, substatus int) bool           { return true }

// LogRecords is a slice of *LogRecord that is sortable
type LogRecords []*LogRecord

// Len implements Sort interface
func (l LogRecords) Len() int { return len(l) }

// Swap implements Sort interface
func (l LogRecords) Swap(i, j int) {
	l[j], l[i] = l[i], l[j]
}

// Less implements Sort interface
func (l LogRecords) Less(i, j int) bool {
	return l[i].DateTime.Before(l[j].DateTime)
}
