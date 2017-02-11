# iislogs

iislog is a tool for fast search events in logs files produced by one or more IIS servers. Searches can be done for:
* user
* errors
* long queries
* url parts

Reports can be limited to a range of dates or hours. 
Static files can also be not included into the report.


## Usage
```
usage: iislog [<flags>] <file>...

an application for searching in IIS logs files,

Flags:
  --help                   Show context-sensitive help (also try --help-long and
                           --help-man).
  --from=DATETIME          get logs from 'DATETIME' UTC
  --to=DATETIME            get logs to 'DATETIME' UTC
  --from-days-ago=DAYS     get log from DAYS ago
  --to-days-ago=DAYS       get log until DAYS before today
  --url=URL ...            Reports lines containing url. Several --url options
                           can be given. Lines are reported whenever one url
                           matches
  --user=USER ...          Reports lines from logged in USER. Several --user
                           options can be given. Lines are reported whenever an
                           user matches
  --errors                 filter logs on protocol errors (4xx and 5xx)
  --hide-assets            hide assets (html,gif,ico,css,jpg,png,js) from result
                           list
  --long-queries=DURATION  show queries longer than 'DURATION'. Accepted values
                           like 200ms, 3s, 1m...

Args:
  <file>   file, path, zip archive


```


## Functionalities
- [X] Limit search between dates time
- [X] Search across several files
- [X] Search in zipped logs
- [X] Search errors 4xx and 5xx
- [X] List all log entries
- [X] Sort by date time entries coming from several servers logs
- [X] List long queries
- [X] List queries from an user
- [X] List queries to an URL
- [ ] Nice unescaped reported queries


