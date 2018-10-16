package catalog

import (
	"errors"
	"fmt"
)

// DefaultFileName ...
const DefaultFileName = "cstore.yml"

// ErrFileRefNotFound ...
var ErrFileRefNotFound = errors.New("file reference not found")

// Catalog ...
type Catalog struct {
	Version string `yaml:"version"`
	Context string `yaml:"context"`

	Files map[string]File `yaml:"files"`
}

// Vault ...
type Vault struct {
	Credentials string `yaml:"credentials,omitempty"`
	Encryption  string `yaml:"encryption,omitempty"`
}

// File ...
type File struct {
	// Path is the location of the file being stored.
	Path string `yaml:"path"`

	// AternatePath is a path used to clone the file to during a restore.
	// This can be used when multiple files tagged differently need to
	// restore to the same path using the same file name.
	AternatePath string `yaml:"alternatePath,omitempty"`

	// Store indicates the remote store the file is stored in.
	Store string `yaml:"store,omitempty"`

	// Encrypted indicates if the file was encrypted before storing.
	Encrypted bool `yaml:"encrypted"`

	// IsRef indicates the file is a linked catalog and not a remotely
	// store file.
	IsRef bool `yaml:"isRef"`

	// IsEnv indicates the file has name value pairs like a .env file.
	IsEnv bool `yaml:"isEnv"`

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

// Key ...
func (f File) Key() string {
	return hashPath(f.Path)
}

// VersionExists ...
func (f File) VersionExists(version string) bool {
	for _, ver := range f.Versions {
		if version == ver {
			return true
		}
	}

	return false
}

// ContextKey ...
func (c Catalog) ContextKey(key string) string {
	return buildKey(c.Context, key)
}

// GetFileNames ...
func (c Catalog) GetFileNames() []string {
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

// GetTaggedPaths ...
func (c Catalog) GetTaggedPaths(tags []string, all bool) []string {
	paths := []string{}
	for _, file := range FilterByTag(c.Files, tags, all) {
		paths = append(paths, file.Path)
	}

	return paths
}

// FilesBy ...
func (c Catalog) FilesBy(args, tags []string, version string) map[string]File {
	return getFiles(c.Files, args, tags, version)
}

func getFiles(files map[string]File, paths []string, tags []string, version string) map[string]File {

	targets := FilterByPath(files, paths)

	targets = FilterByTag(targets, tags, false)

	targets = FilterByVersion(targets, version)

	return targets
}

//FilterByVersion ...
func FilterByVersion(files map[string]File, version string) map[string]File {
	targets := map[string]File{}

	if len(version) == 0 {
		return files
	}

	for key, file := range files {
		if file.VersionExists(version) {
			targets[key] = file
		}
	}

	return targets
}

// FilterByTag ...
func FilterByTag(files map[string]File, tags []string, allTags bool) map[string]File {
	targets := map[string]File{}

	if len(tags) == 0 {
		return files
	}

	for key, file := range files {
		if file.IsRef && len(file.Tags) == 0 {
			targets[key] = file
		}
	}

	for key, file := range files {
		if allTags {
			if areTagsIn(file.Tags, tags) {
				targets[key] = file
			}
		} else {
			for _, tag := range tags {
				if isTagIn(tag, file.Tags) {
					targets[key] = file
				}
			}
		}
	}

	return targets
}

func areTagsIn(tags, tagList []string) bool {
	for _, t := range tags {
		inList := false

		for _, tl := range tagList {
			if tl == t {
				inList = true
			}
		}

		if !inList {
			return false
		}
	}
	return true
}

func isTagIn(tag string, tags []string) bool {
	for _, t := range tags {
		if t == tag {
			return true
		}
	}
	return false
}

// FilterByPath ...
func FilterByPath(files map[string]File, paths []string) map[string]File {
	targets := map[string]File{}

	if len(paths) == 0 {
		return files
	}

	for _, path := range paths {
		for key, file := range files {
			if file.Path == path {
				targets[key] = file
			}
		}
	}

	return targets
}

// Exists ...
func (c *Catalog) Exists(file File) bool {
	_, found := c.Files[file.Key()]

	return found
}

// Update adds the new entry returning the modified
// catalog. (The catalog is not saved at this point.)
func (c *Catalog) Update(newFile File) (File, error) {
	key := newFile.Key()

	// copy previous file info to new file
	if oldFile, found := c.Files[key]; found {
		newFile.Data = oldFile.Data

		if len(newFile.Tags) == 0 {
			newFile.Tags = oldFile.Tags
		}

		if len(newFile.Versions) == 0 {
			newFile.Versions = oldFile.Versions
		}

		if len(newFile.AternatePath) == 0 {
			newFile.AternatePath = oldFile.AternatePath
		}

		if len(newFile.Vaults.Credentials) == 0 {
			newFile.Vaults.Credentials = oldFile.Vaults.Credentials
		}

		if len(newFile.Vaults.Encryption) == 0 {
			newFile.Vaults.Encryption = oldFile.Vaults.Encryption
		}

		if len(newFile.Store) > 0 && newFile.Store != oldFile.Store {
			return newFile, fmt.Errorf("To change store, purge file '%s' and push again.", newFile.Path)
		}

		newFile.Store = oldFile.Store
	}

	c.Files[key] = newFile

	return c.Files[key], nil
}

func buildKey(context, key string) string {
	return fmt.Sprintf("%s/%s", context, key)
}
