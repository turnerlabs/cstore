## Publish GitHub Release ##

NOTE: semver tag format required `v1.0.0-rc`
```bash
$ git tag {{TAG}}
$ git push origin {{TAG}}
$ ./create_darwin_build.sh
```
Once the build is complete the {{TAG}} release will be published to GitHub. 

IMPORTANT: The Mac build is not part of the build process due to required libraries not included on the linux build server. The `create_darwin_build.sh` script must be run on a Mac and manually uploaded to the new GitHub release to complete the process.