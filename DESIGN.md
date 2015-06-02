# gostack mss package

mss stands for measurements, and it's a very simple package for tracing execution of a Go application, measuring and collecting stats for monitoring and review purpose.

## Tracing design

Tags:

* host
* service
* handler
* user_id
* session_id
* request_id

Measurements:

* mss.http
 * url
 * duration
 * content-type
 * status
* mss.database
 * identifier
 * duration
 * query
* mss.rendering
 * file
 * duration
 

## Stories

| Story                                                             | Points |
|-------------------------------------------------------------------|--------|
| Foundation and client for posting measurements to InfluxDB        |      3 |
| Context-based implementation to pass client down for measurements |      3 |
| HTTP measurement                                                  |      1 |
| DB measurement                                                    |      2 |
| Rendering measurement                                             |      1 |
| Grafana 2 dashboard                                               |      3 |

Total of 13 points
