package logging

import (
	"bytes"
	"io"
	"net/http"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// NewLoggingTransport returns a wrapper around http.RoundTripper
// that logs HTTP requests and responses using the `tflog` package.
// The context.Context of the underlying http.Request is passed to the logger.
func NewLoggingTransport(transport http.RoundTripper) http.RoundTripper {
	return &loggingTransport{transport: transport}
}

// loggingTransport is a http.RoundTripper that logs HTTP requests and responses.
type loggingTransport struct {
	transport http.RoundTripper
}

func (t *loggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx := req.Context()

	fields, err := collectRequestFields(req)
	if err != nil {
		tflog.Error(ctx, "Failed to parse request for logging", map[string]interface{}{
			"error": err,
		})
	} else {
		tflog.Debug(ctx, "Sending HTTP Request", fields)
	}

	resp, err := t.transport.RoundTrip(req)
	if err != nil {
		return resp, err
	}

	fields, err = collectResponseFields(resp)
	if err != nil {
		tflog.Error(ctx, "Failed to parse response for logging", map[string]interface{}{
			"error": err,
		})
	} else {
		tflog.Debug(ctx, "Received HTTP Response", fields)
	}
	return resp, nil
}

func collectRequestFields(req *http.Request) (map[string]interface{}, error) {
	fields := make(map[string]interface{})

	fields["http_op"] = "request"
	fields["http_url"] = req.URL.String()
	fields["http_method"] = req.Method

	// Collect request headers
	for name, values := range req.Header {
		if len(values) == 1 {
			fields[name] = values[0]
		} else {
			fields[name] = values
		}
	}

	// Collect the request body
	body, err := bodyFromRequest(req)
	if err != nil {
		return nil, err
	}
	fields["http_req_body"] = body

	return fields, nil
}

func bodyFromRequest(req *http.Request) (string, error) {
	if req.Body == nil {
		return "", nil
	}

	// Read and log the body without consuming it
	var buf bytes.Buffer
	tee := io.TeeReader(req.Body, &buf)

	// Read the body into a byte slice
	bodyBytes, err := io.ReadAll(tee)
	if err != nil {
		return "", err
	}

	// Restore the original request body for the actual request
	req.Body = io.NopCloser(&buf)

	return string(bodyBytes), nil
}

func collectResponseFields(resp *http.Response) (map[string]interface{}, error) {
	fields := make(map[string]interface{})

	fields["http_op"] = "response"
	fields["http_status"] = resp.StatusCode

	// Collect response headers
	for name, values := range resp.Header {
		if len(values) == 1 {
			fields[name] = values[0]
		} else {
			fields[name] = values
		}
	}

	// Collect the response body
	body, err := bodyFromResponse(resp)
	if err != nil {
		return nil, err
	}
	fields["http_resp_body"] = body

	return fields, nil
}

func bodyFromResponse(resp *http.Response) (string, error) {
	if resp.Body == nil {
		return "", nil
	}

	// Read the response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Close the original response body
	err = resp.Body.Close()
	if err != nil {
		return "", err
	}

	// Restore the response body for the client
	resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	return string(bodyBytes), nil
}
