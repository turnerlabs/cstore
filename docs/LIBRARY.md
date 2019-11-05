## How to Load Configuration Using a Library ##

Run `pull` on app start to retrieve configuration. 

<details>
  <summary>Go (direct)</summary>

Requires: `go1.12.0`

```bash
$ export CSTORE_CATALOG="cstore.yml"
```

```go
package main

import (
  "log"
  "os"
  
  "github.com/turnerlabs/cstore"
)

config, err := cstore.Pull(os.Getenv("CSTORE_CATALOG"), 
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
</details>

<details>
  <summary>NodeJS (CLI wrapper)</summary>

### 1. Install cStore at Build Time ###
```bash
$ curl -L -o  /usr/local/bin/cstore https://github.com/turnerlabs/cstore/releases/download/v3.6.0-alpha/cstore_linux_386 && chmod +x /usr/local/bin/cstore
```

### 2. Wrap CLI Pull Command in Module ###

```javascript
'use strict';
const shell = require('shelljs');

function pull(opts) {
  // silent:true prevents output from being logged compromising secrets
  var result = shell.exec('cstore pull -l -g ' + opts.format + ' -t "' + opts.tags.join("&") + '"', {silent:true})

  if (result.code != 0) {
    throw new Error(result.stderr)
  }

  return JSON.parse(result.stdout)
}

module.exports.pull = pull;
```

### 3. Get Configuration ###

```javascript
let config = cstore.pull({
    tags: [process.env.ENV],
    format: "json-object"
})

Object.keys(config).forEach(function (key) {
    console.log(key + "=" + config[key]);
});
```

Need a Node library? Try [cstore-js](https://github.com/shivpatel/cstore-js) instead. 

</details>

<details>
  <summary>C# .Net (CLI wrapper)</summary>

### 1. Install cStore at Build Time ###

```bash
$ curl -L -o  /usr/local/bin/cstore https://github.com/turnerlabs/cstore/releases/download/v3.6.0-alpha/cstore_linux_386 && chmod +x /usr/local/bin/cstore
```

### 2. Wrap CLI Pull Command in Class ###

```C
using System;
using System.Diagnostics;

namespace app
{
    public struct options {
        public string format;

        public string[] tags;

        public string location;
    }


    public static class cstore
    {
        public static string Pull(options opts)
        {
            string cmd = String.Format(@"cstore pull -l");
            
            if (opts.format != null) {
                cmd = String.Format("{0} -g {1}", cmd, opts.format);
            }

            if (opts.tags != null) {
                cmd = String.Format("{0} -t {1}", cmd, String.Join("&", opts.tags));
            }

            var escapedArgs = cmd.Replace("\"", "\\\"");

            var process = new Process()
            {
                StartInfo = new ProcessStartInfo
                {
                    FileName = "/bin/bash",
                    Arguments = $"-c \"{escapedArgs}\"",
                    RedirectStandardOutput = true,
                    WorkingDirectory = opts.location,
                    UseShellExecute = false,
                    CreateNoWindow = false,
                }
            };

            process.Start();
            process.WaitForExit();

            return process.StandardOutput.ReadToEnd();
        }
    }
}
```

### 3. Get Configuration ###

```C
using System;
using System.Collections.Generic;
using Newtonsoft.Json;

namespace app
{
    class Program
    {
        static void Main(string[] args)
        {
            var opts = new options();
            
            opts.format = "json-object";
            opts.tags = new string[] {"dev"};

            var output = cstore.Pull(opts);
        
            var config = JsonConvert.DeserializeObject<Dictionary<string, string>>(output);

            foreach (var pair in config)
                Console.WriteLine("{0}: {1}", pair.Key, pair.Value);
        } 
    }
}
```

</details>