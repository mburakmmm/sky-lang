package skylib

import (
	"bytes"
	"io"
	"net/http"
	"time"
)

// HTTPResponse represents an HTTP response
type HTTPResponse struct {
	StatusCode int
	Headers    map[string]string
	Body       []byte
}

// HTTPRequest represents an HTTP request
type HTTPRequest struct {
	Method  string
	URL     string
	Headers map[string]string
	Body    []byte
}

// HTTPGet performs GET request
func HTTPGet(url string, headers map[string]string) (*HTTPResponse, error) {
	return HTTPRequest{
		Method:  "GET",
		URL:     url,
		Headers: headers,
	}.Send()
}

// HTTPPost performs POST request
func HTTPPost(url string, body []byte, headers map[string]string) (*HTTPResponse, error) {
	return HTTPRequest{
		Method:  "POST",
		URL:     url,
		Headers: headers,
		Body:    body,
	}.Send()
}

// Send sends the HTTP request
func (req HTTPRequest) Send() (*HTTPResponse, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	var bodyReader io.Reader
	if req.Body != nil {
		bodyReader = bytes.NewReader(req.Body)
	}

	httpReq, err := http.NewRequest(req.Method, req.URL, bodyReader)
	if err != nil {
		return nil, err
	}

	// Set headers
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Extract headers
	headers := make(map[string]string)
	for key := range resp.Header {
		headers[key] = resp.Header.Get(key)
	}

	return &HTTPResponse{
		StatusCode: resp.StatusCode,
		Headers:    headers,
		Body:       body,
	}, nil
}

// HTTPServer simple HTTP server
type HTTPServer struct {
	addr    string
	handler func(map[string]interface{}) map[string]interface{}
}

// NewHTTPServer creates HTTP server
func HTTPNewServer(addr string, handler func(map[string]interface{}) map[string]interface{}) *HTTPServer {
	return &HTTPServer{
		addr:    addr,
		handler: handler,
	}
}

// Start starts the server
func (s *HTTPServer) Start() error {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)

		req := map[string]interface{}{
			"method": r.Method,
			"url":    r.URL.String(),
			"path":   r.URL.Path,
			"query":  r.URL.RawQuery,
			"body":   string(body),
		}

		resp := s.handler(req)

		if status, ok := resp["status"].(int); ok {
			w.WriteHeader(status)
		}

		if respBody, ok := resp["body"].(string); ok {
			w.Write([]byte(respBody))
		}
	})

	return http.ListenAndServe(s.addr, nil)
}
