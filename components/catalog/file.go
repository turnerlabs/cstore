package catalog

import (
	"fmt"
	"time"

	yaml "gopkg.in/yaml.v2"

	"github.com/turnerlabs/cstore/components/configure"
)

const name = "pulls.yml"

// FilePurged ...
func (c Catalog) FilePurged(fileName string) error {
	pulls := map[string]time.Time{}

	b, err := configure.Get(name, "")
	if err == nil {
		if err = yaml.Unmarshal(b, &pulls); err != nil {
			return err
		}
	}

	delete(pulls, c.ContextKey(fileName))

	b, err = yaml.Marshal(pulls)
	if err != nil {
		return err
	}

	if err := configure.Update(name, "", b); err != nil {
		return err
	}

	return nil
}

// FilePulled ...
func (c Catalog) FilePulled(fileName, version string, lastPush time.Time) error {
	pulls := map[string]time.Time{}

	b, err := configure.Get(name, "")
	if err == nil {
		if err = yaml.Unmarshal(b, &pulls); err != nil {
			return err
		}
	}

	pulls[c.ContextKey(fileName)] = lastPush
	b, err = yaml.Marshal(pulls)
	if err != nil {
		return err
	}

	if err := configure.Update(name, "", b); err != nil {
		return err
	}

	return nil
}

// FilePulledBefore ...
func (c Catalog) FilePulledBefore(fileName, version string, lastPush time.Time) bool {

	b, err := configure.Get(name, "")
	if err != nil {
		fmt.Print(err)
		return false
	}

	pulls := map[string]time.Time{}
	if err = yaml.Unmarshal(b, &pulls); err != nil {
		fmt.Print(err)
		return false
	}

	if filePulled, found := pulls[c.ContextKey(fileName)]; found {
		return lastPush.After(filePulled)
	}

	return false
}
