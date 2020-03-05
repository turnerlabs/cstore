package s3

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/turnerlabs/cstore/v4/cli/cmd"
	"github.com/turnerlabs/cstore/v4/components/cfg"
	"github.com/turnerlabs/cstore/v4/components/models"
)

const (
	Context     = "automated"
	TestDataDir = "temp"
)

// uncomment when debugging tests locally
//var testWriter = os.Stderr
var testWriter = ioutil.Discard

func TestMain(m *testing.M) {
	setup(m)
	code := m.Run()
	os.Exit(code)
}

func setup(m *testing.M) {
	// create a directory for test data
	if _, err := os.Stat(TestDataDir); os.IsNotExist(err) {
		if err := os.Mkdir(TestDataDir, 0777); err != nil && !os.IsNotExist(err) {
			panic(err)
		}
	}
}

func cleanupOutput(catalog string) {

	if _, err := os.Stat(catalog); !os.IsNotExist(err) {
		opt := cfg.UserOptions{
			Catalog: catalog,
		}

		cmd.Purge(opt, makeIO(testWriter, testWriter, "y"))
	}
}

// Input is processed in the order of the args.
func makeIO(userOutput io.Writer, export io.Writer, args ...interface{}) models.IO {
	input := ""

	for range args {
		input += "%s\n"
	}

	return models.IO{
		UserOutput: userOutput,
		UserInput:  bufio.NewReader(bytes.NewReader([]byte(fmt.Sprintf(input, args...)))),
		Export:     export,
	}

}

type file struct {
	io      models.IO
	data    string
	tags    string
	version string
}
