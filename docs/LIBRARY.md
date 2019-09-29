## How to Load Configuration Using a Library ##

Run this code on app start to retrieve configuration. 

Need a Node library? Try [cstore-js](https://github.com/shivpatel/cstore-js) instead. 

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
