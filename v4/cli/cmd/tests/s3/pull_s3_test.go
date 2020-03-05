package s3

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/turnerlabs/cstore/v4/cli/cmd"
	"github.com/turnerlabs/cstore/v4/components/cfg"
)

//--------------------------------------------------
//- When a pull request is made containing the or
//- operator between two tags, any files containing
//- at least one of the tags should be restored, but
//- files not containing any of the flags shoulbe be
//- ignored.
//--------------------------------------------------
func TestEnsureOnlyFilesContainingAtLeastOneOfTwoTagsAreRestored(t *testing.T) {
	cleanupOutput(fmt.Sprintf("%s.yml", t.Name()))

	// arrange
	expectedTags := "app|local"

	expectedFile1 := fmt.Sprintf("%s/%s1.env", TestDataDir, t.Name())
	expectedFile2 := fmt.Sprintf("%s/%s2.env", TestDataDir, t.Name())
	expectedFile3 := fmt.Sprintf("%s/%s3.env", TestDataDir, t.Name())
	unexpectedFile := fmt.Sprintf("%s/%s4.env", TestDataDir, t.Name())

	files := map[string]file{
		expectedFile1: file{
			io:   makeIO(testWriter, testWriter, os.Getenv("AWS_S3_BUCKET")),
			data: "ENV=dev",
			tags: "app|local",
		},
		expectedFile2: file{
			io:   makeIO(testWriter, testWriter, os.Getenv("AWS_S3_BUCKET")),
			data: "ENV=qa",
			tags: "qa|app",
		},
		expectedFile3: file{
			io:   makeIO(testWriter, testWriter, os.Getenv("AWS_S3_BUCKET")),
			data: "ENV=uat",
			tags: "local",
		},
		unexpectedFile: file{
			io:   makeIO(testWriter, testWriter, os.Getenv("AWS_S3_BUCKET")),
			data: "ENV=uat",
			tags: "uat|other",
		},
	}

	for f, d := range files {
		if err := ioutil.WriteFile(f, []byte(d.data), 0644); err != nil {
			panic(err)
		}
	}

	// act
	pushed := 0
	for f, d := range files {
		if pushed == 0 {
			d.io = makeIO(testWriter, testWriter, fmt.Sprintf("%s-%s", Context, t.Name()), os.Getenv("AWS_S3_BUCKET"))
		}

		opt := cfg.UserOptions{
			Catalog: fmt.Sprintf("%s.yml", t.Name()),
			Paths:   []string{f},
			Tags:    d.tags,
		}

		opt.ParseTags()
		cmd.Push(opt, d.io)

		if err := os.Remove(f); err != nil {
			panic(err)
		}

		pushed++
	}

	opt := cfg.UserOptions{
		Catalog: fmt.Sprintf("%s.yml", t.Name()),
		Tags:    expectedTags,
	}

	opt.ParseTags()
	cmd.Pull(opt.Catalog, opt, makeIO(testWriter, testWriter))

	// assert files with expected tags were pulled down
	if file, err := ioutil.ReadFile(expectedFile1); err != nil {
		t.Errorf("\nEXPECTED: %s found \nACTUAL: file missing", expectedFile1)
	} else if string(file) != files[expectedFile1].data {
		t.Errorf("\nEXPECTED: %s\nACTUAL: %s", files[expectedFile1].data, string(file))
	}

	if file, err := ioutil.ReadFile(expectedFile2); err != nil {
		t.Errorf("\nEXPECTED: %s found \nACTUAL: file missing", expectedFile2)
	} else if string(file) != files[expectedFile2].data {
		t.Errorf("\nEXPECTED: %s\nACTUAL: %s", files[expectedFile2].data, string(file))
	}

	if file, err := ioutil.ReadFile(expectedFile3); err != nil {
		t.Errorf("\nEXPECTED: %s found \nACTUAL: file missing", expectedFile3)
	} else if string(file) != files[expectedFile3].data {
		t.Errorf("\nEXPECTED: %s\nACTUAL: %s", files[expectedFile3].data, string(file))
	}

	// assert file with unexpected tags was not pulled down
	if _, err := ioutil.ReadFile(unexpectedFile); err == nil {
		t.Errorf("\nEXPECTED: file missing\nACTUAL: %s found", unexpectedFile)
	}

	cleanupOutput(fmt.Sprintf("%s.yml", t.Name()))
}

//---------------------------------------------------
//- When a pull request contains the 'i' flag, ensure
//- remote vault secrets are retrieved and injected
//- into a copy of the retreived file.
//---------------------------------------------------
func TestEnsureSecretsAreInjectedIntoFilesDuringPullRequest(t *testing.T) {
	cleanupOutput(fmt.Sprintf("%s.yml", t.Name()))

	// arrange
	expectedFile := fmt.Sprintf("%s/%s1.env", TestDataDir, t.Name())
	expectedFileData := "DB=mongodb://{{dev/user}}:{{dev/password}}@ds111111.mlab.com:111111/app-dev\nAPI_KEY={{dev/key}}"
	expectedFileSecrets := "DB=mongodb://test-user:shh...@ds111111.mlab.com:111111/app-dev\nAPI_KEY=test-api-key"

	files := map[string]file{
		expectedFile: file{
			io:   makeIO(testWriter, testWriter, fmt.Sprintf("%s-%s", Context, t.Name()), os.Getenv("AWS_S3_BUCKET"), os.Getenv("AWS_STORE_KMS_KEY_ID")),
			data: "DB=mongodb://{{dev/user::test-user}}:{{dev/password::shh...}}@ds111111.mlab.com:111111/app-dev\nAPI_KEY={{dev/key::test-api-key}}",
		},
	}

	for f, d := range files {
		if err := ioutil.WriteFile(f, []byte(d.data), 0644); err != nil {
			panic(err)
		}
	}

	// act
	for f, d := range files {
		opt := cfg.UserOptions{
			Catalog:       fmt.Sprintf("%s.yml", t.Name()),
			Paths:         []string{f},
			ModifySecrets: true,
		}

		cmd.Push(opt, d.io)

		if err := os.Remove(f); err != nil {
			panic(err)
		}
	}

	opt := cfg.UserOptions{
		Catalog:       fmt.Sprintf("%s.yml", t.Name()),
		InjectSecrets: true,
	}

	cmd.Pull(opt.Catalog, opt, makeIO(testWriter, testWriter))

	// assert file was pulled down
	if file, err := ioutil.ReadFile(expectedFile); err != nil {
		t.Errorf("\nEXPECTED: %s \nACTUAL: file missing", expectedFile)
		t.Error(err)
	} else if string(file) != expectedFileData {
		t.Errorf("\nEXPECTED: %s\nACTUAL: %s", expectedFileData, string(file))
	}

	// assert file with secrets was pulled down
	secretsFileName := fmt.Sprintf("%s.secrets", expectedFile)
	if file, err := ioutil.ReadFile(secretsFileName); err != nil {
		t.Errorf("\nEXPECTED: %s \nACTUAL: file missing", secretsFileName)
		t.Error(err)
	} else if !strings.Contains(string(file), expectedFileSecrets) {
		t.Errorf("\nEXPECTED: %s\nACTUAL: %s", expectedFileSecrets, string(file))
	}

	cleanupOutput(fmt.Sprintf("%s.yml", t.Name()))
}
