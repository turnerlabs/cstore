package token

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestJSONEnsureTokensAreExtractedFromFile(t *testing.T) {
	// arrange
	file := map[string]interface{}{
		"opt": map[string]interface{}{
			"web": map[string]interface{}{
				"database": map[string]interface{}{
					"url": "http://{{dev/user::app_user}}:{{dev/pass::fhd2!#$^#3sh3e2%k}}@database.com/app",
				},
			},
		},
	}

	data, err := json.Marshal(file)
	if err != nil {
		t.Error(err)
	}

	// act
	tokens, err := Find([]byte(data), "json", true)
	if err != nil {
		t.Error(err)
	}

	// assert
	secret := "dev/opt/web/database/url/user"
	value := "app_user"
	if token, found := tokens[secret]; found {
		if token.Value != value {
			t.Errorf("\nEXPECTED: %s \nACTUAL: %s", value, token.Value)
		}
	} else {
		t.Errorf("\nEXPECTED: %s \nACTUAL: secret missing", secret)
	}

	secret = "dev/opt/web/database/url/pass"
	value = "fhd2!#$^#3sh3e2%k"
	if token, found := tokens[secret]; found {
		if token.Value != value {
			t.Errorf("\nEXPECTED: %s \nACTUAL: %s", value, token.Value)
		}
	} else {
		t.Errorf("\nEXPECTED: %s \nACTUAL: secret missing", secret)
	}
}

func TestEnsureJSONTokensAreReplacedWithValues(t *testing.T) {
	// arrange
	file := map[string]interface{}{
		"opt": map[string]interface{}{
			"web": map[string]interface{}{
				"database": map[string]interface{}{
					"url": "http://{{dev/user}}:{{dev/pass}}@database.com/app",
				},
			},
		},
	}
	expectedFile := map[string]interface{}{
		"opt": map[string]interface{}{
			"web": map[string]interface{}{
				"database": map[string]interface{}{
					"url": "http://app_user:fhd2!#$^#3sh3e2%k@database.com/app",
				},
			},
		},
	}

	data, err := json.Marshal(file)
	if err != nil {
		t.Error(err)
	}

	expectedData, err := json.Marshal(expectedFile)
	if err != nil {
		t.Error(err)
	}

	// act
	b, err := Replace([]byte(data), "json", map[string]Token{
		"1": Token{
			Env:    "dev",
			EnvVar: "opt/web/database/url",
			Prop:   "user",
			Value:  "app_user",
		},
		"2": Token{
			Env:    "dev",
			EnvVar: "opt/web/database/url",
			Prop:   "pass",
			Value:  "fhd2!#$^#3sh3e2%k",
		},
	}, false)
	if err != nil {
		t.Error(err)
	}

	// assert
	if strings.IndexAny(string(b), string(expectedData)) < 0 {
		t.Errorf("\nEXPECTED: %s \nACTUAL: %s", expectedFile, string(b))
	}
}

func TestENVEnsureTokensAreExtractedFromFile(t *testing.T) {
	// arrange
	file := "DB_URL=http://{{dev/user::app_user}}:{{dev/pass::fhd2!#$^#3sh3e2%k}}@database.com/app"

	// act
	tokens, err := Find([]byte(file), "env", true)
	if err != nil {
		t.Error(err)
	}

	// assert
	secret := "dev/db-url/user"
	value := "app_user"
	if token, found := tokens[secret]; found {
		if token.Value != value {
			t.Errorf("\nEXPECTED: %s \nACTUAL: %s", value, token.Value)
		}
	} else {
		t.Errorf("\nEXPECTED: %s \nACTUAL: secret missing", secret)
	}

	secret = "dev/db-url/pass"
	value = "fhd2!#$^#3sh3e2%k"
	if token, found := tokens[secret]; found {
		if token.Value != value {
			t.Errorf("\nEXPECTED: %s \nACTUAL: %s", value, token.Value)
		}
	} else {
		t.Errorf("\nEXPECTED: %s \nACTUAL: secret missing", secret)
	}
}

func TestEnsureENVTokensAreReplacedWithValues(t *testing.T) {
	// arrange
	file := "DB_URL=http://{{dev/user}}:{{dev/pass}}@database.com/app"
	expectedFile := "DB_URL=http://app_user:fhd2!#$^#3sh3e2%k@database.com/app"

	data, err := json.Marshal(file)
	if err != nil {
		t.Error(err)
	}

	// act
	b, err := Replace([]byte(data), "json", map[string]Token{
		"1": Token{
			Env:    "dev",
			EnvVar: "db-url",
			Prop:   "user",
			Value:  "app_user",
		},
		"2": Token{
			Env:    "dev",
			EnvVar: "db-url",
			Prop:   "pass",
			Value:  "fhd2!#$^#3sh3e2%k",
		},
	}, false)
	if err != nil {
		t.Error(err)
	}

	// assert
	if strings.IndexAny(string(b), expectedFile) < 0 {
		t.Errorf("\nEXPECTED: %s \nACTUAL: %s", expectedFile, string(b))
	}
}
