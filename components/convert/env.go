package convert

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/subosito/gotenv"
)

// ToJSONObjectFormat ...
func ToJSONObjectFormat(file []byte) (bytes.Buffer, error) {
	reader := bytes.NewReader(file)
	pairs := gotenv.Parse(reader)

	var buff bytes.Buffer

	env := map[string]string{}

	for key, value := range pairs {
		env[key] = value
	}

	b, err := json.MarshalIndent(env, "", "    ")
	if err != nil {
		return buff, err
	}

	_, err = buff.Write(b)
	if err != nil {
		return buff, err
	}

	return buff, nil
}

// ToENVFileFormat ...
func ToENVFileFormat(file []byte) (bytes.Buffer, error) {
	var buff bytes.Buffer

	env := map[string]string{}

	if err := json.Unmarshal(file, &env); err != nil {
		return buff, err
	}

	for k, v := range env {
		if _, err := buff.WriteString(fmt.Sprintf("%s=%s\n", k, v)); err != nil {
			return buff, err
		}
	}

	return buff, nil
}
