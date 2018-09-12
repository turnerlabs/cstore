# Habor Store #

The Harbor store pushes and pulls environment variables to and from Harbor by linking a `.env` file to a Harbor container during the initial push. Only environment variables pushed using cStore can be pulled or deleted through cStore allowing cStore to ignore environment variables on the same container in Harbor that were added through the GUI.

## Environment Variables ##

### Prefixing ###

Each environment variable pushed is prefixed in the `cstore.yml` file with `ENV_` to allow cstore to distinguish between other data related to the pushed file. For example, an environment variable called `URL` would be pushed to Harbor as `URL`, but will be stored in the `cstore.yml` file as `ENV_URL`.

### Types ###

All environment variables pushed to Harbor will be defaulted to type `hidden` for highest level of security, but they can be changed after the initial push by editing the `cstore.yml` file and pushing again. Available types are `basic` or `hidden`. 

### Example cstore.yml ###

```
version: v1
context: 01655ed0-61b8-4da7-8d50-a266ce4330de
files:
  0b288e8e36e43f9172058245c0d18c72:
    path: environments/dev/.env
    data:
      ENV_URL: hidden
```
