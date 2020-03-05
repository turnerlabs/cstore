package contract

import (
	"errors"
	"time"

	"github.com/turnerlabs/cstore/v4/components/catalog"
	"github.com/turnerlabs/cstore/v4/components/cfg"
	"github.com/turnerlabs/cstore/v4/components/models"
)

// IStore is a persistence abstraction facilitating the
// implementation of different file storage solutions capable
// of saving, retrieving, and deleting the contents of a file
// using a unique file key.
//
// To add a new store, create a struct implementing this interface
// in a separate file under the stores folder using the naming
// convention `{type}_store.go`.
type IStore interface {

	// Name is unique store identifier that is use as a command line
	// flag.
	Name() string

	// SupportsFeature should return true or false depending on which storage
	// feature is passed. An example would be support for versioning.
	SupportsFeature(feature string) bool

	// SupportsFileType should return true or false depending on which file
	// type is passed. Examples would be support for .env or .json files.
	SupportsFileType(fileType string) bool

	// Description provides details on how to use the store and is
	// displayed on the command line when store details are requested.
	Description() string

	// Pre is executed before any other store actions and ensures
	// required data for the store to operate exists and is valid.
	// It should return nil to indicate the store is ready to be used.
	//
	// Often store authentication, validation, and defaulting are
	// performed in Pre.
	//
	// If the store needs to persist information in the catalog, the
	// "file.Data" map can be used. Since the file is passed by ref,
	// there is no need to return the file for data to be saved.
	//
	// "access" is used to get secrets required to access the store.
	// It is common to set the vault on the struct to allow other
	// methods access to them.
	//
	// "uo" specifies the users intention when pushing or pulling a
	// file if the store can provide additional features.
	//
	// "io" contains readers and writers that should be used when
	// displaying instructions to or reading data from the command
	// line.
	//
	// "clog" contains the context id, unique identifier, for the yml
	// file created when content is pushed to the store. It can be
	// used to ensure unique to file names for this context.
	//
	// "error" should return nil if the operation was successful.
	Pre(clog catalog.Catalog, file *catalog.File, access IVault, uo cfg.UserOptions, io models.IO) error

	// Push is called when a file needs to be remotely stored.
	// The file should be stored using a combination of the catalog's
	// context that could be retreived from the struct and the file's
	// path that can be hashed if too long.
	//
	// If the store needs to persist information in the catalog, the
	// "file.Data" map can be used. Since the file is passed by ref,
	// there is no need to return the file for data to be saved.
	//
	// "version" contains the version that the file contents should be
	// stored under, but does not need to be used if the "Supports"
	// function does not indicate the store supports versioning.
	// Versioning is the ability for a store copy the file contents
	// and store/retrieve it separately from the working copy.
	//
	// "error" should return nil if the operation was successful.
	Push(file *catalog.File, fileData []byte, version string) error

	// Pull is called when a file needs to be retrieved from the remote store.
	//
	// "version" contains the version of the file contents that should
	// be retrieved, but does not need to be used if the "Supports"
	// function does not indicate the store supports versioning.
	//
	// The "[]byte"" array should be the contents of the retrieved file.
	//
	// "Attributes" should return the time the file was last updated.
	//
	// "error" should return nil if the operation was successful.
	Pull(file *catalog.File, version string) ([]byte, Attributes, error)

	// Purge is called when a file needs to be deleted from the remote store.
	//
	// "version" contains the version of the file contents that should
	// be deleted, but does not need to be used if the "Supports"
	// function does not indicate the store supports versioning. Purge
	// should not delete the working copy of the file if len(version) > 0.
	//
	// "error" should return nil if the operation was successful.
	Purge(file *catalog.File, version string) error

	// Changed is called to determine when a file last changed. This is
	// used to prompt the user to overrite if desired.
	//
	// "time.Time" should return time.Time{} when file is not found.
	//
	// "error" should return nil if the operation was successful.
	Changed(file *catalog.File, fileData []byte, version string) (time.Time, error)
}

// ErrStoreNotFound is returned when the store is not implemented.
var ErrStoreNotFound = errors.New("store not found")

// Attributes ...
type Attributes struct{}
