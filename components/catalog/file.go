package catalog

import (
	"errors"
	"time"

	"github.com/turnerlabs/cstore/components/local"
	"github.com/turnerlabs/cstore/components/logger"
	yaml "gopkg.in/yaml.v2"
)

const name = "state.yml"

// RemoveRecords ...
func (c Catalog) RemoveRecords(fileName string) error {
	pulls := map[string]State{}

	b, err := local.Get(name, "")
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

	return local.Update(name, "", b)
}

// RecordPull ...
func (c Catalog) RecordPull(fileName string, lastPull time.Time, version string) error {
	if lastPull.IsZero() {
		return errors.New("invalid time")
	}

	pulls := map[string]State{}

	b, err := local.Get(name, "")
	if err == nil {
		if err = yaml.Unmarshal(b, &pulls); err != nil {
			return err
		}
	}

	s := State{
		Pulled:  lastPull.UTC(),
		Version: version,
	}

	pulls[c.ContextKey(fileName)] = s
	b, err = yaml.Marshal(pulls)
	if err != nil {
		return err
	}

	return local.Update(name, "", b)
}

// IsCurrent ...
func (f File) IsCurrent(lastChange time.Time, context string) (bool, string) {

	if lastChange.IsZero() {
		return true, ""
	}

	b, err := local.Get(name, "")
	if err != nil {
		logger.L.Print(err)
		return false, ""
	}

	pulls := map[string]State{}
	if err = yaml.Unmarshal(b, &pulls); err != nil {
		logger.L.Print(err)
		return false, ""
	}

	if filePulled, found := pulls[f.ContextKey(context)]; found {
		return filePulled.Pulled.After(lastChange) || filePulled.Pulled.Equal(lastChange), filePulled.Version
	}

	return false, ""
}

// State contains information that is relavant to the pulled file.
type State struct {
	Pulled  time.Time
	Version string `yaml:"version,omitempty"`
}
