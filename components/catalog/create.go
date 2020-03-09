package catalog

import (
	"github.com/turnerlabs/cstore/v4/components/cfg"
	"github.com/turnerlabs/cstore/v4/components/models"
	"github.com/turnerlabs/cstore/v4/components/prompt"
)

func create(io models.IO) Catalog {
	val := prompt.GetValFromUser("Context", prompt.Options{
		Description:  "The project name categorizing the remotely stored files. This gives context to all files in this catalog and is often used as a prefix in the remote store. To avoid overriding existing data in the remote store, ensure context is unique.",
		DefaultValue: getContext(),
	}, io)

	return Catalog{
		Version: cfg.Version[0:2],
		Context: val,
		Files:   map[string]File{},
	}
}
