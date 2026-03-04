// name: jwt
// description: JSON Web Token utilities
// author: roturbot
// requires: encoding/json, encoding/base64, strings, time

type JWT struct{}

func (JWT) Encode(header any, payload any, secret any) string {
	headerStr := OSLtoString(header)
	payloadStr := OSLtoString(payload)
	secretStr := OSLtoString(secret)

	headerBytes := []byte(headerStr)
	secretBytes := []byte(secretStr)

	header64 := btoa(string(headerBytes))
	payloadBytes := []byte(payloadStr)
	payload64 := btoa(string(payloadBytes))

	signatureInput := header64 + "." + payload64
	signature := OSLSHA256(signatureInput + "." + secretStr)

	return signatureInput + "." + signature
}

type jwtHeader struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

type jwtPayload map[string]any

func (JWT) Sign(claims map[string]any, secret any, expiresIn any) string {
	secretStr := OSLtoString(secret)
	duration := int(OSLcastNumber(expiresIn))

	header := jwtHeader{
		Alg: "HS256",
		Typ: "JWT",
	}

	headerJSON, _ := json.Marshal(header)
	payloadJSON, _ := json.Marshal(claims)

	header64 := btoa(string(headerJSON))
	payload64 := btoa(string(payloadJSON))

	signatureInput := header64 + "." + payload64
	signature := OSLSHA256(secretStr + "." + signatureInput)

	token := signatureInput + "." + signature
	return token
}

func (JWT) SignWithExpiry(claims map[string]any, secret any, expiresIn any) string {
	duration := int(OSLcastNumber(expiresIn))
	expiry := time.Now().Add(time.Duration(duration) * time.Second).Unix()

	claims["exp"] = expiry
	claims["iat"] = time.Now().Unix()

	return jwt.Sign(claims, secret, expiresIn)
}

func (JWT) Verify(token any, secret any) map[string]any {
	tokenStr := OSLtoString(token)
	secretStr := OSLtoString(secret)

	parts := strings.Split(tokenStr, ".")
	if len(parts) != 3 {
		return map[string]any{"valid": false, "error": "invalid_token_format"}
	}

	header64, payload64, signature := parts[0], parts[1], parts[2]

	signatureInput := header64 + "." + payload64
	expectedSignature := OSLSHA256(secretStr + "." + signatureInput)

	if len(signature) != len(expectedSignature) {
		return map[string]any{"valid": false, "error": "invalid_signature"}
	}

	sigMatch := true
	for i := 0; i < len(signature); i++ {
		if i < len(expectedSignature) && signature[i] != expectedSignature[i] {
			sigMatch = false
		} else if i >= len(expectedSignature) {
			sigMatch = false
		}
	}

	if !sigMatch {
		return map[string]any{"valid": false, "error": "invalid_signature"}
	}

	headerJSON, err := atob(header64)
	if err != nil || headerJSON == "" {
		return map[string]any{"valid": false, "error": "invalid_header"}
	}

	payloadJSON, err := atob(payload64)
	if err != nil || payloadJSON == "" {
		return map[string]any{"valid": false, "error": "invalid_payload"}
	}

	var payload map[string]any
	json.Unmarshal([]byte(payloadJSON), &payload)

	if exp, ok := payload["exp"].(float64); ok {
		if int64(exp) < time.Now().Unix() {
			return map[string]any{"valid": false, "error": "token_expired"}
		}
	}

	result := map[string]any{
		"valid":   true,
		"header":  OSLJsonParse(headerJSON),
		"payload": payload,
	}
	return result
}

func (JWT) Decode(token any) map[string]any {
	tokenStr := OSLtoString(token)

	parts := strings.Split(tokenStr, ".")
	if len(parts) != 3 {
		return map[string]any{"error": "invalid_token_format"}
	}

	header64, payload64 := parts[0], parts[1]

	headerJSON, _ := atob(header64)
	payloadJSON, _ := atob(payload64)

	var header, payload map[string]any
	json.Unmarshal([]byte(headerJSON), &header)
	json.Unmarshal([]byte(payloadJSON), &payload)

	return map[string]any{
		"header":  header,
		"payload": payload,
	}
}

func (JWT) GetClaim(token any, claim any) any {
	tokenStr := OSLtoString(token)
	claimStr := OSLtoString(claim)

	result := jwt.Decode(tokenStr)
	if payload, ok := result["payload"].(map[string]any); ok {
		return payload[claimStr]
	}
	return nil
}

func (JWT) IsExpired(token any) bool {
	tokenStr := OSLtoString(token)

	result := jwt.Decode(tokenStr)
	if payload, ok := result["payload"].(map[string]any); ok {
		if exp, ok := payload["exp"].(float64); ok {
			return int64(exp) < time.Now().Unix()
		}
	}
	return true
}

func (JWT) Refresh(token any, secret any, expiresIn any) string {
	tokenStr := OSLtoString(token)

	result := jwt.Verify(tokenStr, secret)
	if verifyResult, ok := result["valid"].(bool); ok && verifyResult {
		if payload, ok := result["payload"].(map[string]any); ok {
			delete(payload, "exp")
			delete(payload, "iat")
			delete(payload, "jti")
			return jwt.SignWithExpiry(payload, secret, expiresIn)
		}
	}
	return ""
}

var jwt = JWT{}
