# checkPhish

A simple Go client for the CheckPhish API.

Go to checkphish.ai to get an apikey.

# Usage: 
go build main.go
./main evil.com apikey

## Index

* [Constants](#constants)
* [func NewClient](#func-newclient)
* [type CPClient](#type-cpclient)
  * [func (c *CPClient) NewSearchJob](#func-cpclient-newsearchjob)
  * [func (c *CPClient) NewStatusJob](#func-cpclient-newstatusjob)
  * [func (c *CPClient) GetResultJob](#func-cpclient-getresultjob)
* [func cpErrCodes](#func-cperrcodes)
* [type CPScanResponse](#type-cpscanresponse)
* [type CPStatusResponse](#type-cpstatusresponse)
* [type CPDoneResponse](#type-cpdoneresponse)

## Constants

```go
const baseURL = "https://developers.checkphish.ai/api"
```

## func NewClient

```go
func NewClient(apiKey string) *CPClient
```

NewClient returns a CPClient configured with a 30-second timeout.

## type CPClient

```go
type CPClient struct {
    // contains filtered or unexported fields
}
```

CPClient is a client for the CheckPhish API.

### func (*CPClient) NewSearchJob

```go
func (c *CPClient) NewSearchJob(ctx context.Context, url string) (*CPScanResponse, error)
```

NewSearchJob submits a URL for scanning and returns the job ID.

### func (*CPClient) NewStatusJob

```go
func (c *CPClient) NewStatusJob(ctx context.Context, id string) (*CPStatusResponse, error)
```

NewStatusJob checks the current status of a previously submitted job.

### func (*CPClient) GetResultJob

```go
func (c *CPClient) GetResultJob(ctx context.Context, id string) (*CPDoneResponse, error)
```

GetResultJob retrieves the final result for a completed job.

## func cpErrCodes

```go
func cpErrCodes(code int) error
```

cpErrCodes maps HTTP status codes from the CheckPhish API to errors.

## type CPScanResponse

```go
type CPScanResponse struct {
    Jobid     string `json:"jobID"`
    Timestamp int64  `json:"timestamp"`
}
```

CPScanResponse is returned when a new scan job is submitted.

## type CPStatusResponse

```go
type CPStatusResponse struct {
    JobID  string `json:"job_id"`
    URL    string `json:"url"`
    Status string `json:"status"`
}
```

CPStatusResponse contains the current status of a scan job.

## type CPDoneResponse

```go
type CPDoneResponse struct {
    JobID       string `json:"job_id"`
    Status      string `json:"status"`
    URL         string `json:"url"`
    URLSha256   string `json:"url_sha256"`
    Disposition string `json:"disposition"`
    Brand       string `json:"brand"`
    Error       bool   `json:"error"`
}
```
