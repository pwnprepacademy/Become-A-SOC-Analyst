package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const baseURL = "https://developers.checkphish.ai/api"

type CPClient struct {
	hc     *http.Client
	apiKey string
}

func NewClient(apiKey string) *CPClient {
	return &CPClient{
		hc: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				TLSHandshakeTimeout: 10 * time.Second,
				// Prefer keeping TLS verification enabled.
				// Only disable if you have a specific reason.
				// TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
		apiKey: apiKey,
	}
}

func (c *CPClient) do(ctx context.Context, method, path string, headers map[string]string) ([]byte, int, error) {
	req, err := http.NewRequestWithContext(ctx, method, baseURL+path, nil)
	if err != nil {
		return nil, 0, err
	}

	req.Header.Set("apikey", c.apiKey)
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := c.hc.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}

	if err := cpErrCodes(resp.StatusCode); err != nil {
		return body, resp.StatusCode, err
	}

	return body, resp.StatusCode, nil
}

func (c *CPClient) NewSearchJob(ctx context.Context, url string) (*CPScanResponse, error) {
	headers := map[string]string{
		"urlInfo":  fmt.Sprintf(`{"url":%q}`, url),
		"insights": "true",
	}

	body, _, err := c.do(ctx, http.MethodPost, "/neo/scan", headers)
	if err != nil {
		return nil, err
	}

	var resp CPScanResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *CPClient) NewStatusJob(ctx context.Context, id string) (*CPStatusResponse, error) {
	headers := map[string]string{"jobID": id}

	body, _, err := c.do(ctx, http.MethodPost, "/neo/scan/status", headers)
	if err != nil {
		return nil, err
	}

	var resp CPStatusResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *CPClient) GetResultJob(ctx context.Context, id string) (*CPDoneResponse, error) {
	// Note: Original code used the status endpoint for results too.
	// Confirm with CheckPhish docs if a different endpoint is needed.
	headers := map[string]string{"jobID": id}

	body, _, err := c.do(ctx, http.MethodPost, "/neo/scan/status", headers)
	if err != nil {
		return nil, err
	}

	var resp CPDoneResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s <URL> <apikey>\n", os.Args[0])
		os.Exit(1)
	}

	url := os.Args[1]
	apiKey := os.Args[2]

	client := NewClient(apiKey)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Submit scan
	scan, err := client.NewSearchJob(ctx, url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "scan failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Job submitted: %s\n", scan.Jobid)

	// Poll until done or timeout
	const maxAttempts = 10
	for i := 0; i < maxAttempts; i++ {
		status, err := client.NewStatusJob(ctx, scan.Jobid)
		if err != nil {
			fmt.Fprintf(os.Stderr, "status check failed: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Status: %s\n", status.Status)

		if status.Status != "PENDING" {
			result, err := client.GetResultJob(ctx, scan.Jobid)
			if err != nil {
				fmt.Fprintf(os.Stderr, "result fetch failed: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Result: %+v\n", result)
			return
		}

		select {
		case <-ctx.Done():
			fmt.Fprintln(os.Stderr, "timed out waiting for result")
			os.Exit(1)
		case <-time.After(15 * time.Second):
			// continue polling
		}
	}

	fmt.Fprintln(os.Stderr, "max polling attempts reached")
	os.Exit(1)
}

func cpErrCodes(code int) error {
	switch code {
	case http.StatusOK:
		return nil
	case http.StatusBadRequest:
		return fmt.Errorf("400 request error")
	case http.StatusUnauthorized:
		return fmt.Errorf("401 authentication failure")
	case http.StatusNotFound:
		return fmt.Errorf("404 endpoint does not exist")
	case http.StatusConflict:
		return fmt.Errorf("409 invalid operation")
	case http.StatusInternalServerError:
		return fmt.Errorf("500 internal server error")
	case http.StatusServiceUnavailable:
		return fmt.Errorf("503 feature disabled")
	default:
		return fmt.Errorf("unexpected status code: %d", code)
	}
}

type CPScanResponse struct {
	Jobid     string `json:"jobID"`
	Timestamp int64  `json:"timestamp"`
}

type CPStatusResponse struct {
	JobID  string `json:"job_id"`
	URL    string `json:"url"`
	Status string `json:"status"`
}

type CPDoneResponse struct {
	JobID       string `json:"job_id"`
	Status      string `json:"status"`
	URL         string `json:"url"`
	URLSha256   string `json:"url_sha256"`
	Disposition string `json:"disposition"`
	Brand       string `json:"brand"`
	Error       bool   `json:"error"`
}
