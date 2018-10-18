## User Configuration ##

If `$HOME/.cstore/user.yml` file is created, CStore will use user defaults when specific flags are not specified.

```
# default store to use when pushing a file the first time
store: aws-s3 

# override the file specific credentials and encryption vaults locally.
credentials: env
encryption: osx-keychain

# set a custom file for cstore.yml files
file: mystore.yml
```