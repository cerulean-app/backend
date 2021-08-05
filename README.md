# cerulean-backend

The backend for Cerulean.

## Setup

If you would like to setup your own Cerulean backend, first compile Cerulean using `go build` and then run `./cerulean-backend` (`.\cerulean-backend.exe` on Windows) after creating the `config.json` file. The `config.json` should look like this:

```json
{
  "port": 7292,
  "mongoUri": "<MongoDB connection URI>"
}
```
