### Linking Catalogs ###

Sometimes, it can be useful to restore multiple catalogs with a single command. For this purpose, catalogs can be linked to a parent catalog which allows restoration of the parent catalog to restore child catalogs at the same time.
```bash
$ cstore push {{path}}/cstore.yml
$ cstore pull
```
Pushing a catalog file, `cstore.yml`, will link the pushed catalog to the parent catalog.

Purging of a parent catalog will not purge the contents of linked child catalogs, but listing a parent's contents will include the contents of the linked children.

If a linked catalog is tagged, the linked catalog's files will only be accessed when the tag is used. Otherwise, tags will flow down and be applied to the linked catalog's files.