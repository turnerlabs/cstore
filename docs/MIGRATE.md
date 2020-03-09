### Migrate from V1 to V3 ###

To migrate to v2, pull all files locally using v1 and push them back up using v3 before purging v1 files from the remote store.

```bash
$ cd {{PROJECT}}
$ cstore pull
$ mv cstore.yml cstore.yml.v1
$ mv  /usr/local/bin/cstore /usr/local/bin/cstore-v1 
$ sudo curl -L -o /usr/local/bin/cstore https://github.com/turnerlabs/cstore/releases/download/v3.8.3-alpha/cstore_darwin_amd64 && sudo chmod +x /usr/local/bin/cstore
$ cstore push // files pulled from old version of cStore
```

Upgrade all instances of cStore before executing the next v1 clean up steps.

When pulling configuration into a docker container, the `docker-entrypoint.sh` script should be updated by removing the `-c` flag which is no longer needed to authenticate with AWS and adding the `-l` flag which removes terminal formatting making output friendlier for logs.

```bash
$ cstore-v1 purge -f cstore.yml.v1
$ rm /usr/local/bin/cstore-v1
```
