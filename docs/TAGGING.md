### Tagging Files / Configuration ###

A single catalog may store multiple files for any given folder context, but a complete restore, `pull`, of the catalog may not always be needed and specifying each file by path is tedious. For this purpose, files can be tagged when pushed using a `|` delimited list of tags `$ cstore push {file} -t "dev|qa"`. At this point, files can be listed, purged, or retrieved by tags. For example, `$ cstore pull -t dev` only restores files containing a `dev` tag.

The commands `list`, `purge`, and `pull` accepts a list of tags deliniated by a `|` or `&`. If any listed tags are deliniated by a `|`, files matching any tag in the list will be retrieved. If all tags in the list are deliniated by an `&`, only files containing all tags in the list will be retrieved.

Multiple tags should be encapsulated by quotes. (i.e. `"dev|secure|vscode"`)

If no tags are specified on the initial push, tags will be parsed from the folder location of the file. For example, a path like `service/dev/.env` would create tags `service` and `dev` and store them with the file in the catalog.

When pushing without specifying tags, the file will keep the tags from the last push. When tags are pushed, the files previous tags will be replaced with the new tags.