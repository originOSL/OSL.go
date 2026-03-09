// name: requests
// description: HTTP utilities
// author: Mist
// requires: net/http, encoding/json, net/url as OSLurl

type HTTP struct {
	Client *http.Client
}

func extractOptions(data map[string]any) (headers map[string]string, body OSLio.Reader, params map[string]string) {
	headers = make(map[string]string)
	params = make(map[string]string)

	if data == nil {
		return
	}

	_, hasHeaders := data["headers"]
	_, hasParams := data["params"]
	isStructured := hasHeaders || hasParams

	if isStructured {
		if raw, ok := data["headers"]; ok {
			if hmap, ok := raw.(map[string]any); ok {
				for k, v := range hmap {
					headers[k] = OSLtoString(v)
				}
			}
		}

		if raw, ok := data["params"]; ok {
			if pmap, ok := raw.(map[string]any); ok {
				for k, v := range pmap {
					params[k] = OSLtoString(v)
				}
			}
		}
	} else {
		for k, v := range data {
			if k == "body" {
				continue
			}
			headers[k] = OSLtoString(v)
		}
	}

	if raw, ok := data["body"]; ok {
		switch v := raw.(type) {
		case string:
			body = bytes.NewReader([]byte(v))
		case []byte:
			body = bytes.NewReader(v)
		case map[string]any:
			body = bytes.NewReader([]byte(JsonStringify(v)))
			if headers["Content-Type"] == "" {
				headers["Content-Type"] = "application/json"
			}
		case []any:
			body = bytes.NewReader([]byte(JsonStringify(v)))
			if headers["Content-Type"] == "" {
				headers["Content-Type"] = "application/json"
			}
		default:
			buf, _ := json.Marshal(v)
			body = bytes.NewReader(buf)
			if headers["Content-Type"] == "" {
				headers["Content-Type"] = "application/json"
			}
		}
	}

	return
}

func (h *HTTP) doRequest(method, rawURL string, data map[string]any) map[string]any {
	headers, body, params := extractOptions(data)

	out := map[string]any{
		"headers": nil,
		"body":    nil,
		"raw":     nil,
		"status":  0,
		"success": false,
	}

	if len(params) > 0 {
		parsed, err := OSLurl.Parse(rawURL)
		if err != nil {
			return out
		}
		q := parsed.Query()
		for k, v := range params {
			q.Set(k, v)
		}
		parsed.RawQuery = q.Encode()
		rawURL = parsed.String()
	}

	req, err := http.NewRequest(method, rawURL, body)
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

	respBody, err := OSLio.ReadAll(resp.Body)
	if err != nil {
		return out
	}

	out["body"] = respBody
	return out
}

func (h *HTTP) Request(method any, url any, data ...map[string]any) map[string]any {
	var m map[string]any
	if len(data) > 0 {
		m = data[0]
	}
	return h.doRequest(strings.ToUpper(OSLtoString(method)), OSLtoString(url), m)
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
	headers, _, params := extractOptions(m)

	rawURL := OSLtoString(url)
	if len(params) > 0 {
		parsed, err := OSLurl.Parse(rawURL)
		if err != nil {
			return out
		}
		q := parsed.Query()
		for k, v := range params {
			q.Set(k, v)
		}
		parsed.RawQuery = q.Encode()
		rawURL = parsed.String()
	}

	req, err := http.NewRequest(http.MethodHead, rawURL, nil)
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
	out["status"] = resp.StatusCode
	out["success"] = true
	return out
}

var requests = &HTTP{Client: http.DefaultClient}