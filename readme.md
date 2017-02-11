# iislog

iislog is a tool for fast search events in logs files produced by one or more IIS servers. Searches can be done for:
* user
* errors
* long queries
* url parts

Reports can be limited to a range of dates or hours. 
Static files can be ignored in the result.

Result is send to console, in CSV format. First line has header.
Example:

```
isslogs.exe --from-days-ago=12 --hide-assets --errors Logs-IIS\*IIS*.zip
date;status;s-ip;cs-username;cs-uri-stem;cs-uri-query;time-taken(ms);time-taken;status-label
"2017-01-31 09:08:40";500.0;10.30.136.200;DOMAIN\user;/myapp/;"|121|800a0046|File:___Permission_denied__Error_opening_log_file_C:\Windows\TEMP\API.Log";2246;2.246s;"Module or ISAPI error occurred"
"2017-01-31 10:34:43";500.0;10.30.136.200;DOMAIN\user;/myapp/;"|121|800a0046|File:___Permission_denied__Error_opening_log_file_C:\Windows\TEMP\API.Log";1216;1.216s;"Module or ISAPI error occurred"
"2017-01-31 10:35:03";404.0;10.30.136.200;-;/myapps/;"-";93;93ms;"Not found"
"2017-01-31 10:35:14";500.0;10.30.136.200;DOMAIN\user;/myapp/;"|121|800a0046|File:___Permission_denied__Error_opening_log_file_C:\Windows\TEMP\API.Log";889;889ms;"Module or ISAPI error occurred"
```

## Usage
```
usage: iislog [<flags>] <file>...

a tool for searching in IIS logs files.

  Author jfc@responsiveconsulting.fr

Flags:
  --help                   Show context-sensitive help (also try --help-long and
                           --help-man).
  --from=DATETIME          get logs from 'DATETIME' UTC
  --to=DATETIME            get logs to 'DATETIME' UTC
  --from-days-ago=DAYS     get logs from DAYS ago
  --to-days-ago=DAYS       get logs until DAYS before today
  --since=DURATION         get logs since DURATION. DURATION can be like 2s,
                           24h...
  --url=URL ...            Reports lines containing url. Several --url options
                           can be given. Lines are reported whenever one url
                           matches
  --user=USER ...          Reports lines from authenticated USER. Several --user
                           options can be given. Lines are reported whenever an
                           user matches
  --errors                 filter logs on protocol errors (4xx and 5xx)
  --hide-assets            hide assets (html,gif,ico,css,jpg,png,js) from result
                           list
  --long-queries=DURATION  show queries longer than 'DURATION'. Accepted values
                           like 200ms, 3s, 1m...

Args:
  <file>  file, path, zip archive


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


