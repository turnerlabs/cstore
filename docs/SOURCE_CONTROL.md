## Set Up Source Control ##

Source Control requires a source control repository.

After creating a file locally, `$ cstore push .env` will register the file with the catalog, `cstore.yml` file, expecting the file to be checked into source control. Functionality like pulling and secret injection from a remote vault will then be available.

### Versioning Configuration ###

Versioning is currently not necesary. Tipically a File Reference store should be checked into version control.