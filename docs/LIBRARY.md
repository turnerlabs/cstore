## How to Load Configuration Using a Library ##

Run `pull` on app start to retrieve configuration. 

### Using Library (Go) ### 

Requires: `go1.12.0`

```go
package main

import (
	"log"

	"github.com/turnerlabs/cstore"
)

config, err := cstore.Pull("cstore.yml", 
    cstore.Options{
        Tags:          []string{"dev"},
        Version:       "v1.8.0-rc",
        InjectSecrets: true,
})

if err != nil {
    log.Fatal(err)
}

for k, v := range config {
    log.Printf("%s=%s\n", k, v)
}
```

### Using CLI (javascript) ###

This method should work with any language. Below is a javascript example.

1. Install cStore at build time
```bash
$ curl -L -o  /usr/local/bin/cstore https://github.com/turnerlabs/cstore/releases/download/v3.4.0-alpha/cstore_linux_386 && chmod +x /usr/local/bin/cstore
```

2. Wrap CLI pull command
```javascript

'use strict';
const shell = require('shelljs');

function pull(cstorePath, tags) {
  // silent:true prevents output from being logged compromising secrets
  var result = shell.exec('./' + cstorePath + ' pull -l -g json-object -t "' + tags + '"', {silent:true})

  if (result.code != 0) {
    throw new Error(result.stderr)
  }

  return JSON.parse(result.stdout)
}

module.exports.pull = pull;
```

Need a Node library? Try [cstore-js](https://github.com/shivpatel/cstore-js) instead. 

