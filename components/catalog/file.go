package catalog

import (
	"errors"
	"time"

	"github.com/turnerlabs/cstore/components/local"
	"github.com/turnerlabs/cstore/components/logger"
	yaml "gopkg.in/yaml.v2"
)

const name = "pulls.yml"

// RemoveRecords ...
func (c Catalog) RemoveRecords(fileName string) error {
	pulls := map[string]time.Time{}

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
func (c Catalog) RecordPull(fileName string, lastPull time.Time) error {
	if lastPull.IsZero() {
		return errors.New("invalid file pulled time")
	}

	pulls := map[string]time.Time{}

	b, err := local.Get(name, "")
	if err == nil {
		if err = yaml.Unmarshal(b, &pulls); err != nil {
			return err
		}
	}

	pulls[c.ContextKey(fileName)] = lastPull.UTC()
	b, err = yaml.Marshal(pulls)
	if err != nil {
		return err
	}

	return local.Update(name, "", b)
}

// IsCurrent ...
func (f File) IsCurrent(lastChange time.Time, context string) bool {

	if lastChange.IsZero() {
		return true
	}

	b, err := local.Get(name, "")
	if err != nil {
		logger.L.Print(err)
		return false
	}

	pulls := map[string]time.Time{}
	if err = yaml.Unmarshal(b, &pulls); err != nil {
		logger.L.Print(err)
		return false
	}

	if filePulled, found := pulls[f.ContextKey(context)]; found {
		return filePulled.After(lastChange) || filePulled.Equal(lastChange)
	}

	return false
}
