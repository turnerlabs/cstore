package s3

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/turnerlabs/cstore/cli/cmd"
	"github.com/turnerlabs/cstore/components/cfg"
)

//---------------------------------------------------
//- When multiple files are pushed to a remote store
//- individually, they should be able to be retrieved
//- individually.
//---------------------------------------------------
func TestEnsureMultipleFilesCanBeRetrievedAfterPushingIndividually(t *testing.T) {
	cleanupOutput(fmt.Sprintf("%s.yml", t.Name()))

	// arrange
	files := map[string]file{
		fmt.Sprintf("%s/%s1.env", TestDataDir, t.Name()): file{
			io:   makeIO(testWriter, testWriter, os.Getenv("AWS_S3_BUCKET")),
			data: "ENV=test1",
		},
		fmt.Sprintf("%s/%s2.env", TestDataDir, t.Name()): file{
			io:   makeIO(testWriter, testWriter, os.Getenv("AWS_S3_BUCKET")),
			data: "ENV=test2",
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
		}

		cmd.Push(opt, d.io)

		if err := os.Remove(f); err != nil {
			panic(err)
		}

		cmd.Pull(opt.Catalog, opt, d.io)

		pushed++
	}

	// assert
	for f, d := range files {
		file, err := ioutil.ReadFile(f)
		if err != nil {
			t.Errorf("\nEXPECTED: %s\nACTUAL: file missing", f)
			t.Error(err)
		}

		if string(file) != d.data {
			t.Errorf("\nEXPECTED: %s\nACTUAL: %s", d.data, string(file))
		}
	}

	cleanupOutput(fmt.Sprintf("%s.yml", t.Name()))
}

//---------------------------------------------------
//- When a file is versioned, the version should be
//- restored when pulled specifically.
//---------------------------------------------------
func TestEnsurePushedVersionCanBeRetrieved(t *testing.T) {
	cleanupOutput(fmt.Sprintf("%s.yml", t.Name()))

	// arrange
	expectedFile := fmt.Sprintf("%s/%s1.env", TestDataDir, t.Name())

	files := map[string]file{
		expectedFile: file{
			io:      makeIO(testWriter, testWriter, os.Getenv("AWS_S3_BUCKET")),
			data:    "VER=1",
			version: "v1.0.0",
		},
		fmt.Sprintf("%s/%s2.env", TestDataDir, t.Name()): file{
			io:      makeIO(testWriter, testWriter, os.Getenv("AWS_S3_BUCKET")),
			data:    "VER=2",
			version: "v2.0.0",
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
			Version: d.version,
		}

		cmd.Push(opt, d.io)

		if err := os.Remove(f); err != nil {
			panic(err)
		}

		pushed++
	}

	opt := cfg.UserOptions{
		Catalog: fmt.Sprintf("%s.yml", t.Name()),
		Version: files[expectedFile].version,
	}

	cmd.Pull(opt.Catalog, opt, makeIO(testWriter, testWriter))

	// assert
	if file, err := ioutil.ReadFile(expectedFile); err != nil {
		t.Errorf("\nEXPECTED: %s\nACTUAL: file missing", files[expectedFile].version)
		t.Error(err)
	} else if string(file) != files[expectedFile].data {
		t.Errorf("\nEXPECTED: %s\nACTUAL: %s", files[expectedFile].data, string(file))
	}

	cleanupOutput(fmt.Sprintf("%s.yml", t.Name()))
}

//---------------------------------------------------
//- When a file is tagged during a push, a pull using
//- the same tag should retrieve it.
//---------------------------------------------------
func TestEnsureTaggedFileCanBeRetrievedByTag(t *testing.T) {
	cleanupOutput(fmt.Sprintf("%s.yml", t.Name()))

	// arrange
	expectedFile := fmt.Sprintf("%s/%s1.env", TestDataDir, t.Name())
	unexpectedFile := fmt.Sprintf("%s/%s2.env", TestDataDir, t.Name())

	files := map[string]file{
		expectedFile: file{
			io:   makeIO(testWriter, testWriter, os.Getenv("AWS_S3_BUCKET")),
			data: "ENV=dev",
			tags: "dev",
		},
		unexpectedFile: file{
			io:   makeIO(testWriter, testWriter, os.Getenv("AWS_S3_BUCKET")),
			data: "ENV=qa",
			tags: "qa",
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
		Tags:    files[expectedFile].tags,
	}

	opt.ParseTags()
	cmd.Pull(opt.Catalog, opt, makeIO(testWriter, testWriter))

	// assert file with expected tag was pulled down
	if file, err := ioutil.ReadFile(expectedFile); err != nil {
		t.Errorf("\nEXPECTED: %s\nACTUAL: file missing", expectedFile)
		t.Error(err)
	} else if string(file) != files[expectedFile].data {
		t.Errorf("\nEXPECTED: %s\nACTUAL: %s", files[expectedFile].data, string(file))
	}

	// assert file with unexpected tag was not pulled down
	if _, err := ioutil.ReadFile(unexpectedFile); err == nil {
		t.Errorf("\nEXPECTED: file missing\nACTUAL: %s", unexpectedFile)
		t.Error(err)
	}

	cleanupOutput(fmt.Sprintf("%s.yml", t.Name()))
}
