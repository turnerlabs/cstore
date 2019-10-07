package store

import (
	"errors"
	"os"
	"time"

	"github.com/turnerlabs/cstore/components/catalog"
	"github.com/turnerlabs/cstore/components/cfg"
	"github.com/turnerlabs/cstore/components/contract"
	localFile "github.com/turnerlabs/cstore/components/file"
	"github.com/turnerlabs/cstore/components/models"
)

// SourceControlStore ...
type SourceControlStore struct {
	clog catalog.Catalog
}

// Name ...
func (s SourceControlStore) Name() string {
	return "source-control"
}

// SupportsFeature ...
func (s SourceControlStore) SupportsFeature(feature string) bool {
	switch feature {
	case VersionFeature:
		return false
	case SourceControlFeature:
		return true
	default:
		return false
	}
}

// SupportsFileType ...
func (s SourceControlStore) SupportsFileType(fileType string) bool {
	return true
}

// Description ...
func (s SourceControlStore) Description() string {
	return `
	detail: https://github.com/turnerlabs/cstore/blob/master/docs/SOURCE_CONTROL.md
`
}

// Pre ...
func (s *SourceControlStore) Pre(clog catalog.Catalog, file *catalog.File, access contract.IVault, uo cfg.UserOptions, io models.IO) error {
	s.clog = clog

	return nil
}

// Push ...
func (s SourceControlStore) Push(file *catalog.File, fileData []byte, version string) error {

	if len(fileData) == 0 {
		return errors.New("empty file")
	}

	return nil
}

// Pull ...
func (s SourceControlStore) Pull(file *catalog.File, version string) ([]byte, contract.Attributes, error) {

	b, err := localFile.GetBy(s.clog.GetFullPath(file.Path))
	if err != nil {
		return b, contract.Attributes{}, err
	}

	return b, contract.Attributes{}, nil
}

// Purge ...
func (s SourceControlStore) Purge(file *catalog.File, version string) error {
	return os.Remove(s.clog.GetFullPath(file.Path))
}

// Changed ...
func (s SourceControlStore) Changed(file *catalog.File, fileData []byte, version string) (time.Time, error) {
	return time.Time{}, nil
}

func init() {
	s := new(SourceControlStore)
	stores[s.Name()] = s
}
