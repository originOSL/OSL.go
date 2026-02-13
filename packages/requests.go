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
			headers[k] = OSLtoString(v)
		}
	}
	return headers, body
}

func (h *HTTP) doRequest(method, url string, data map[string]any) map[string]any {
	headers, body := extractHeadersAndBody(data)

	out := make(map[string]any)
	out["headers"] = nil
	out["body"] = nil
	out["raw"] = nil
	out["status"] = 0
	out["success"] = false

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return out
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := h.Client.Do(req)
	if err != nil {
		return out
	}
	defer resp.Body.Close()

	respHeaders := make(map[string]any)
	for k, v := range resp.Header {
		if len(v) == 1 {
			respHeaders[k] = v[0]
		} else {
			respHeaders[k] = v
		}
	}

	out["status"] = resp.StatusCode
	out["headers"] = respHeaders
	out["raw"] = resp
	out["success"] = true

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return out
	}

	out["body"] = respBody

	return out
}

func (h *HTTP) Get(url any, data ...map[string]any) map[string]any {
	var m map[string]any
	if len(data) > 0 {
		m = data[0]
	}
	return h.doRequest(http.MethodGet, OSLtoString(url), m)
}

func (h *HTTP) Post(url any, data map[string]any) map[string]any {
	return h.doRequest(http.MethodPost, OSLtoString(url), data)
}

func (h *HTTP) Put(url any, data map[string]any) map[string]any {
	return h.doRequest(http.MethodPut, OSLtoString(url), data)
}

func (h *HTTP) Patch(url any, data map[string]any) map[string]any {
	return h.doRequest(http.MethodPatch, OSLtoString(url), data)
}

func (h *HTTP) Delete(url any, data ...map[string]any) map[string]any {
	var m map[string]any
	if len(data) > 0 {
		m = data[0]
	}
	return h.doRequest(http.MethodDelete, OSLtoString(url), m)
}

func (h *HTTP) Options(url any, data ...map[string]any) map[string]any {
	var m map[string]any
	if len(data) > 0 {
		m = data[0]
	}
	return h.doRequest(http.MethodOptions, OSLtoString(url), m)
}

func (h *HTTP) Head(url any, data ...map[string]any) map[string]any {
	var m map[string]any
	if len(data) > 0 {
		m = data[0]
	}
	out := map[string]any{"success": false}
	headers, _ := extractHeadersAndBody(m)
	req, err := http.NewRequest(http.MethodHead, OSLtoString(url), nil)
	if err != nil {
		return out
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := h.Client.Do(req)
	if err != nil {
		return out
	}
	out["status"] = resp.StatusCode
	defer resp.Body.Close()
	out["success"] = true
	return out
}

var requests = &HTTP{Client: http.DefaultClient}
