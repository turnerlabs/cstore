package catalog

import (
	"fmt"
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

// GhostFile is the file created in place of pushed files that
// allows cstore commands to be executed in the same directory
// as the pushed files even if the cstore.yml is in a different
// directory.
const GhostFile = ".cstore"

// WriteGhost ...
func WriteGhost(path string, g Ghost) error {
	d, err := yaml.Marshal(&g)
	if err != nil {
		return err
	}

	comment := `# Ghost replacement files are created in the directories of remotely stored files. 
# These files make it possible to run cStore commands from the local directory of remotely 
# stored files without being in the same directory as the catalog file. To learn more, 
# visit https://github.com/turnerlabs/cstore/blob/master/docs/GHOST.md.
`
	d = append([]byte(comment), d...)

	return ioutil.WriteFile(fmt.Sprintf("%s/%s", path, GhostFile), d, 0644)
}

// GetGhost ...
func GetGhost() (Ghost, error) {

	g := Ghost{}

	b, err := ioutil.ReadFile(GhostFile)
	if err != nil {
		return g, err
	}

	if err = yaml.Unmarshal(b, &g); err != nil {
		return g, err
	}

	return g, nil
}

// Ghost ...
type Ghost struct {
	Location string
}
