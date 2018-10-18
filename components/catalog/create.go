package catalog

import (
	"github.com/turnerlabs/cstore/components/cfg"
	"github.com/turnerlabs/cstore/components/models"
	"github.com/turnerlabs/cstore/components/prompt"
)

func create(io models.IO) Catalog {
	opt := prompt.Options{
		Description:  "The project name categorizing the remotely stored files. This gives context to all files in this catalog.",
		DefaultValue: getContext(),
	}
	val := prompt.GetValFromUser("Context", opt, io)

	return Catalog{
		Version: cfg.Version[0:2],
		Context: val,
		Files:   map[string]File{},
	}
}
