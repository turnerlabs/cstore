package catalog

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/turnerlabs/cstore/v4/components/path"
)

// DefaultFileName ...
const DefaultFileName = "cstore.yml"

// ErrFileRefNotFound ...
var ErrFileRefNotFound = errors.New("file reference not found")

// ErrSecretNotFound ...
var ErrSecretNotFound = errors.New("not found")

// Version ...
type Version struct {
	Version string `yaml:"version"`
	Context string `yaml:"context"`
}

// Catalog ...
type Catalog struct {
	CWD string `yaml:"-"`

	Version string `yaml:"version"`
	Context string `yaml:"context"`

	Files map[string]File `yaml:"files"`
}

// ToFile ...
func (c Catalog) ToFile() FileCatalog {
	n := FileCatalog{
		Version: c.Version,
		Context: c.Context,
		Files:   []File{},
	}

	for _, f := range c.Files {
		n.Files = append(n.Files, f)
	}

	return n
}

// FileCatalog ...
type FileCatalog struct {
	Version string `yaml:"version"`
	Context string `yaml:"context"`

	Files []File `yaml:"files"`
}

// ToBusiness ...
func (c FileCatalog) ToBusiness(location string) Catalog {
	n := Catalog{
		CWD:     location,
		Version: c.Version,
		Context: c.Context,
		Files:   map[string]File{},
	}

	for _, f := range c.Files {
		n.Files[f.Key()] = f
	}

	return n
}

// Vault ...
type Vault struct {
	Access  string `yaml:"access,omitempty"`
	Secrets string `yaml:"secrets,omitempty"`
}

// File ...
type File struct {
	// Path is the location of the file being stored.
	Path string `yaml:"path"`

	// AlternatePath is a path used to clone the file to during a restore.
	// This can be used when multiple files tagged differently need to
	// restore to the same path using the same file name.
	AlternatePath string `yaml:"alternatePath,omitempty"`

	// Store indicates the remote store the file is stored in.
	Store string `yaml:"store,omitempty"`

	// IsRef indicates the file is a linked catalog and not a remotely
	// store file.
	IsRef bool `yaml:"isRef,omitempty"`

	// DeleteAfterPush instructs the local file to be deleted after changes
	// have been pushed to the remote store to protect secrets
	DeleteAfterPush bool `yaml:"deleteAfterPush,omitempty"`

	// Type indicates what type of contents are in the file like env or json.
	Type string `yaml:"type"`

	// Data is additional info a store may need when restoring a file.
	Data map[string]string `yaml:"data,omitempty"`

	// Tags allow files to be grouped; so, they can be listed, purged,
	// and restored in a single command.
	Tags []string `yaml:"tags,omitempty"`

	// Vaults defines where the credential and encryption information
	// is stored.
	Vaults Vault `ymal:"vaults,omitempty"`

	// Versions stores an identifier for user versioned copies of the data.
	Versions []string `ymal:"versions,omitempty"`
}

// ActualPath ...
func (f File) ActualPath() string {
	return path.SubstituteTokens(f.Path)
}

// Key ...
func (f File) Key() string {
	return hashPath(f.Path)
}

// ContextKey ...
func (f File) ContextKey(context string) string {
	return buildKey(context, hashPath(f.ActualPath()))
}

// SupportsSecrets ...
func (f File) SupportsSecrets() bool {
	supportedTypes := []string{"env", "json"}

	for _, st := range supportedTypes {
		if strings.ToLower(f.Type) == st {
			return true
		}
	}

	return false
}

// SupportsConfig ...
func (f File) SupportsConfig() bool {
	supportedTypes := []string{"env"}

	for _, st := range supportedTypes {
		if strings.ToLower(f.Type) == st {
			return true
		}
	}

	return false
}

// HasStore ...
func (f File) HasStore() bool {
	return len(f.Store) > 0
}

// CleanWorking ...
func (f File) CleanWorking() bool {
	return f.Store != "source-control"
}

// AddData ...
func (f *File) AddData(data map[string]string) {
	if f.Data == nil {
		f.Data = data
	} else {
		for key, entry := range data {
			f.Data[key] = entry
		}
	}
}

// Missing ...
func (f File) Missing(version string) bool {
	for _, ver := range f.Versions {
		if version == ver {
			return false
		}
	}

	return true
}

// Name ...
func (f File) Name() string { return "*.yml" }

// Set ...
func (f *File) Set(contextID, group, prop, value string) error {
	if f.Data == nil {
		f.Data = map[string]string{}
	}

	f.Data[f.BuildKey(contextID, group, prop)] = value

	return nil
}

// Delete ...
func (f *File) Delete(contextID, group, prop string) error {
	return errors.New("not implemented")
}

// BuildKey ...
func (f File) BuildKey(contextID, group, prop string) string {
	return strings.ToUpper(prop)
}

// Get ...
func (f File) Get(contextID, group, prop string) (string, error) {
	if value, found := f.Data[f.BuildKey(contextID, group, prop)]; found {
		return value, nil
	}
	return "", ErrSecretNotFound
}

// ContextKey ...
func (c Catalog) ContextKey(key string) string {
	return buildKey(c.Context, key)
}

// GetPaths ...
func (c Catalog) GetPaths() []string {
	paths := []string{}

	for _, file := range c.Files {
		paths = append(paths, file.Path)
	}

	return paths
}

// GetTagsBy ...
func (c Catalog) GetTagsBy(path string) []string {
	for _, file := range c.Files {
		if file.Path == path {
			return file.Tags
		}
	}

	return []string{}
}

// GetPathsBy ...
func (c Catalog) GetPathsBy(tags []string, all bool) []string {
	paths := []string{}
	for _, file := range keepFilesWithTags(c.Files, tags, all) {
		paths = append(paths, file.Path)
	}

	return paths
}

// GetDataByStore ...
func (c Catalog) GetDataByStore(store, key, defaultValue string) string {
	for _, f := range c.Files {
		if v, exists := f.Data[key]; exists && (store == f.Store || len(store) == 0) {
			return v
		}
	}

	return defaultValue
}

// GetDataByVault ...
func (c Catalog) GetDataByVault(vault, key, defaultValue string) string {
	for _, f := range c.Files {
		if v, exists := f.Data[key]; exists && (vault == f.Vaults.Secrets || len(vault) == 0) {
			return v
		}
	}

	return defaultValue
}

// Location ...
func (c Catalog) Location() string {
	return location(c.CWD)
}

func location(cwd string) string {
	if len(cwd) == 0 {
		return ""
	}

	loc := ""
	for folder := 0; folder <= strings.Count(strings.Trim(cwd, "/"), "/"); folder++ {
		loc = fmt.Sprintf("%s../", loc)
	}

	return loc
}

// GetFullPath ...
func (c Catalog) GetFullPath(path string) string {
	if len(c.CWD) > 0 {
		return fmt.Sprintf("%s%s", c.Location(), path)
	}
	return path
}

// FilesBy ...
func (c Catalog) FilesBy(paths, tags []string, allTags bool, version string) map[string]File {

	filtered := keepFilesWithPaths(c.Files, paths)

	filtered = keepFilesWithTags(filtered, tags, allTags)

	filtered = keepFilesWithVersion(filtered, version)

	for key, file := range c.Files {
		if file.IsRef {
			filtered[key] = file
		}
	}

	return filtered
}

// AnyFilesIn ...
func (c Catalog) AnyFilesIn(dir string) bool {
	for _, f := range c.Files {
		if path.RemoveFileName(f.ActualPath()) == dir {
			return true
		}
	}
	return false
}

// Exists ...
func (c *Catalog) Exists(file File) bool {
	_, found := c.Files[file.Key()]

	return found
}

// LookupEntry ...
func (c *Catalog) LookupEntry(path string, data []byte) (File, bool, error) {

	if file, found := c.Files[hashPath(path)]; found && !file.IsRef {
		return file, true, nil
	}

	file, err := createNew(path, data)

	return file, false, err
}

// UpdateEntry adds the new entry returning the modified
// catalog. (The catalog is not saved at this point.)
func (c *Catalog) UpdateEntry(newFile File) error {
	key := newFile.Key()

	if oldFile, found := c.Files[key]; found {
		if len(newFile.Store) > 0 && newFile.Store != oldFile.Store {
			return fmt.Errorf("AreadyStoredException: Purge %s from %s before pushing to %s", newFile.ActualPath(), oldFile.Store, newFile.Store)
		}
	}

	c.Files[key] = newFile

	return nil
}

func buildKey(context, key string) string {
	return fmt.Sprintf("%s/%s", context, key)
}

func createNew(path string, data []byte) (File, error) {

	isRef, err := IsOne(data)
	if err != nil {
		return File{}, err
	}

	file := File{
		Path:   path,
		IsRef:  isRef,
		Type:   strings.TrimLeft(filepath.Ext(path), "."),
		Vaults: Vault{},
	}

	return file, nil
}
