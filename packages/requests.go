// name: requests
// description: HTTP utilities
// author: Mist
// requires: net/http, encoding/json, io

type HTTP struct {
	Client *http.Client
}

func extractHeadersAndBody(data map[string]any) (headers map[string]string, body io.Reader) {
	headers = make(map[string]string)
	if data != nil {
		if raw, ok := data["body"]; ok {
			switch v := raw.(type) {
			case string:
				body = bytes.NewReader([]byte(v))
			case []byte:
				body = bytes.NewReader(v)
			case map[string]any:
				body = bytes.NewReader([]byte(JsonStringify(v)))
				headers["Content-Type"] = "application/json"
			default:
				buf, _ := json.Marshal(v)
				body = bytes.NewReader(buf)
				headers["Content-Type"] = "application/json"
			}
		}

		for k, v := range data {
			if k == "body" {
				continue
			}
			headers[k] = OSLcastString(v)
		}
	}
	return headers, body
}

func (h *HTTP) doRequest(method, url string, data map[string]any) string {
	headers, body := extractHeadersAndBody(data)

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return ""
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := h.Client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}

	return string(respBody)
}

func (h *HTTP) Get(url string, data ...map[string]any) string {
	var m map[string]any
	if len(data) > 0 {
		m = data[0]
	}
	return h.doRequest(http.MethodGet, url, m)
}

func (h *HTTP) Post(url string, data map[string]any) string {
	return h.doRequest(http.MethodPost, url, data)
}

func (h *HTTP) Put(url string, data map[string]any) string {
	return h.doRequest(http.MethodPut, url, data)
}

func (h *HTTP) Patch(url string, data map[string]any) string {
	return h.doRequest(http.MethodPatch, url, data)
}

func (h *HTTP) Delete(url string, data ...map[string]any) string {
	var m map[string]any
	if len(data) > 0 {
		m = data[0]
	}
	return h.doRequest(http.MethodDelete, url, m)
}

func (h *HTTP) Options(url string, data ...map[string]any) string {
	var m map[string]any
	if len(data) > 0 {
		m = data[0]
	}
	return h.doRequest(http.MethodOptions, url, m)
}

func (h *HTTP) Head(url string, data ...map[string]any) string {
	var m map[string]any
	if len(data) > 0 {
		m = data[0]
	}
	headers, _ := extractHeadersAndBody(m)
	req, err := http.NewRequest(http.MethodHead, url, nil)
	if err != nil {
		return ""
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := h.Client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	return resp.Status
}

var requests = &HTTP{Client: http.DefaultClient}
