package interop

import (
	"bytes"
	"encoding/json"
)

func Serialize(input any) (string, error) {
	jsonBytes, err := json.Marshal(input)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

func Deserialize(input string, result any) error {
	dec := json.NewDecoder(bytes.NewReader([]byte(input)))
	// make sure to throw errors incase unknown fields are passed, do not silently ignore this
	// as it is most likely a sign of buggy interface code
	dec.DisallowUnknownFields()
	return dec.Decode(result)
}
