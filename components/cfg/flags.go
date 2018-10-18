package cfg

import (
	"fmt"
	"strings"
)

// UserOptions ...
type UserOptions struct {
	Store                string
	Tags                 string
	AllTags              bool
	TagList              []string
	Paths                []string
	Version              string
	AlternateRestorePath string
	ModifySecrets        bool
	InjectSecrets        bool
	DeleteLocalFiles     bool
	ExportEnv            bool
	Catalog              string
	AccessVault          string
	SecretsVault         string
	ViewTags             bool
	ViewVersions         bool
	Prompt               bool
	Format               Formatting
}

// Formatting ...
type Formatting struct {
	Bold    string
	UnBold  string
	Red     string
	Blue    string
	NoColor string
}

// AddPaths ...
func (o *UserOptions) AddPaths(paths []string) {
	o.Paths = []string{}

	for _, p := range paths {
		if strings.Index(p, "./") == 0 {
			p = strings.Replace(p, "./", "", 1)
		}

		o.Paths = append(o.Paths, p)
	}
}

// GetPaths ...
func (o *UserOptions) GetPaths(CWD string) []string {

	if len(CWD) == 0 {
		return o.Paths
	}

	if strings.LastIndex(CWD, "/") != len(CWD)-1 {
		CWD += "/"
	}

	paths := []string{}
	for _, path := range o.Paths {
		paths = append(paths, fmt.Sprintf("%s%s", CWD, path))
	}

	return paths
}

// TagsFrom ...
func (o UserOptions) TagsFrom(path string) []string {
	options := strings.Split(path, "/")

	if len(options) <= 1 {
		return []string{}
	}

	return options[0 : len(options)-1]
}

// ParseTags ...
func (o *UserOptions) ParseTags() {
	const and = "&"
	const or = "|"

	sep := and
	o.AllTags = true

	if strings.IndexAny(o.Tags, or) > -1 {
		sep = or

		o.AllTags = false

		o.Tags = strings.Replace(o.Tags, and, or, -1)
	}

	o.TagList = strings.Split(o.Tags, sep)

	if len(o.TagList) == 1 && o.TagList[0] == "" {
		o.TagList = []string{}
	}
}
