package adapter

import (
	"encoding/base64"
)

func Base64Decode(s string) string {
	base64Str, err := base64.StdEncoding.DecodeString(s)
	if err == nil {
		return string(base64Str)
	}

	base64Str, err = base64.RawURLEncoding.DecodeString(s)
	if err == nil {
		return string(base64Str)
	}

	base64Str, err = base64.RawStdEncoding.DecodeString(s)
	if err == nil {
		return string(base64Str)
	}

	return s
}
