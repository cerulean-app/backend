# cerulean-backend

The backend for Cerulean.

## API Documentation

Cerulean provides a REST API for you to be able to write client applications and integrations around. [Click here to view the documentation.](https://github.com/cerulean-app/backend/blob/main/API.md)

## Setup

If you would like to setup your own Cerulean backend, first compile Cerulean using `go build` and then run `./cerulean-backend` (`.\cerulean-backend.exe` on Windows) after creating the `config.json` file. The `config.json` should look like this:

```json
{
  "port": 7292,
  "mongoUri": "<MongoDB connection URI>"
}
```
