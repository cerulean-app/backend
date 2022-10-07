# Cerulean REST API Documentation

Welcome to the documentation for Cerulean's REST API! The REST API is open for everyone to use and write their own custom clients if you feel that our clients are inadequate! We also maintain this documentation to help ease the development of our own clients. This documentation is kept in-tree and updated with the latest commit so that you can go through git history for older versions. [Report any issues with the documentation here.](https://github.com/cerulean-app/backend/issues)

**This REST API is still a work in progress!** Some endpoints and features have not been finalised. If you are writing a Cerulean client, you will need to keep track of the current API until beta versions are released. Additionally, some of these endpoints have not yet been implemented in the back-end.

- `POST /todos/order`
- Repeated todo items API is WIP.
- `POST /register` email verification.
- `POST /resendverifyemail`
- `POST /forgotpassword`
- `POST /verifyuser`

All dates sent to and fro from the REST API are encoded as ISO 8601 strings as emitted by `Date#toISOString` in JavaScript. All request and response bodies, if present, are formatted in JSON.

## [Authentication Scheme](#authentication-scheme)

For authentication, the Cerulean REST API requires an `Authorization` header or a `cerulean_token` cookie to be sent with every request, containing a token which is sent to the client after logging in using the [POST /login](#post-login) endpoint. The [POST /logout](#post-logout) endpoint can be used to invalidate the user's token. The token returned expires after 6 months. Your client should be well equipped to handle token expiries.

### [Extra Authentication Info](#extra-authentication-info)

The user's password can be changed using the [POST /changepassword](#post-changepassword) endpoint. The minimum password length is 8 characters for security reasons. Calling this endpoint logs the user out everywhere except their current session. A user can be registered using the [POST /register](#post-register) endpoint, after which they will be sent an email containing a link to a webpage with a token in the query string, which upon loading will call [POST /verifyuser](#post-verifyuser) to activate the account with the token in the query string. This token has an expiry date of 24 hours, and can be resent by calling [POST /resendverifyemail](#post-resendverifyemail).

The user's account can be deleted with [POST /deleteaccount](#post-deleteaccount). This deletion is permanent, and cannot be undone. Hence, this endpoint should be treated with caution.

## [Syncing Todo Lists](#syncing-todo-lists)

If you are writing a client, and your client goes offline, there are 2 ways to ensure that your client can continue to work offline without messing up any data on the back-end that may be more up to date. It is highly advisable to follow these guidelines. One way to cache all todos on the client, and display them in a read-only mode until an internet connection is available again. However, this is not an ideal user experience.

The ideal way is to cache the todos on the client and create an array of todo IDs which your local client has deleted. When you reconnect to the back-end, you must call the [GET /todos](#get-todos) endpoint to get the latest todos from the server. You can then compare this with your own cache by taking the response, deleting todos from it that your client deleted, modifying todos sharing the same ID in your local cache and the response, and adding todos present in your cache which are not present on the server. Additionally, you can compare the order of the local cache and server response and calculate an appropriate order for the merged lists. Once you have done the comparison, you can send the required DELETE/PATCH/POST requests to the server to sync your local cache with the server, and then overwrite your local cache with [GET /todos](#get-todos).

We may eventually provide a POST /sync endpoint, which would take all todos on the client as well as the client's old cache, merge them with the back-end's copy, and send back a merged list of todos to the client. This would reduce the amount of client-side logic and provide resistance against network failures, which may cause unexpected behaviour.

## [Errors](#errors)

Each endpoint may return certain errors, which have been documented in the description for their response. In addition to the documented errors, every endpoint could return a 5xx HTTP error code which should be handled correctly by the client, and 405 Method Not Allowed and 400 Bad Request if the client is sending invalid requests which do not comply with the parameters. Apart from `/login` and `/register`, all endpoints require the `cerulean_token` cookie (set by `/login` if `cookie` query param is not `false`) or an `Authorization` header, containing a valid session access token, else you will receive 401 Unauthorized.

When there is an error, the server will reply with an object containing `error`, a string with an error message in English (UK). This message is intended to be shown directly to the user. We will eventually add a `code` or `lang` field to the response with corresponding documentation for proper i18n support.

## [POST /register](#post-register)

Register a new Cerulean account. Bienvenue !

### <a name="post-register-parameters">[Parameters](#post-register-parameters)</a>

| Name       | Type    | In    | Description                                                                 |
| ---------- | ------- | ----- | --------------------------------------------------------------------------- |
| `username` | string  | body  | A username (a-zA-Z0-9_) to register with, not already used. Min length: 4   |
| `email`    | string  | body  | A valid email to register with, not already registered.                     |
| `password` | string  | body  | The password to register with. Minimum length: 8.                           |
| `cookie`   | boolean | query | Optional: Set to `false` to avoid getting `Set-Cookie: cerulean_token=`     |

### <a name="post-register-response">[Response](#post-register-response)</a>

Possible errors include 409 Conflict if someone has an account with the existing username and email, and 400 Bad Request if the username, email or password fail validation.

```json
{"token":"JRPnrZPzeb8hi+RigUYZjIBWg4N1hImlI+AwKkfi4fk"}
```

## [POST /login](#post-login)

Log into the Cerulean API and retrieve a token.

### <a name="post-login-parameters">[Parameters](#post-login-parameters)</a>

| Name       | Type    | In    | Description                 |
| ---------- | ------- | ----- | --------------------------- |
| `username` | string  | body  | The username to login with. |
| `password` | string  | body  | The password to login with. |
| `cookie`   | boolean | query | Optional: Set to `false` to avoid getting `Set-Cookie: cerulean_token=` |

### <a name="post-login-response">[Response](#post-login-response)</a>

Possible errors include 401 Unauthorized if your username or password is invalid.

```json
{"token":"JRPnrZPzeb8hi+RigUYZjIBWg4N1hImlI+AwKkfi4fk"}
```

## [POST /logout](#post-logout)

Logout and invaliate the current token.

### <a name="post-logout-parameters">[Parameters](#post-logout-parameters)</a>

| Name | Type | In | Description |
| ---- | ---- | -- | ----------- |
| N/A

### <a name="post-logout-response">[Response](#post-logout-response)</a>

```json
{"success":true}
```

## [POST /changepassword](#post-changepassword)

Change your current user's password. This also invalidates all tokens except your current one.

### <a name="post-changepassword-parameters">[Parameters](#post-changepassword-parameters)</a>

| Name              | Type   | In   | Description                                         |
| ----------------- | ------ | ---- | --------------------------------------------------- |
| `currentPassword` | string | body | The old password.                                   |
| `newPassword`     | string | body | The new password you wish to set. Minimum length: 8 |

### <a name="post-changepassword-response">[Response](#post-changepassword-response)</a>

Possible errors include 400 Bad Request if your new password is less than 8 characters and 401 Unauthorized if the provided current password is incorrect.

```json
{"success":true}
```

## [POST /deleteaccount](#post-deleteaccount)

Delete your account. Au revoir. This is irreversible and logs you out. If writing a client, make sure to cover any calls to this with a big warning dialog.

### <a name="post-deleteaccount-parameters">[Parameters](#post-deleteaccount-parameters)</a>

| Name | Type | In | Description |
| ---- | ---- | -- | ----------- |
| N/A

### <a name="post-deleteaccount-response">[Response](#post-deleteaccount-response)</a>

```json
{"success":true}
```

## [GET /todos](#get-todos)

Get all of the user's todo items. [Read the parameters for POST /todo to help understand the response of this endpoint fully.](#post-todo-parameters) `id`, `createdAt` and `updatedAt` are created by the server and cannot be edited directly.

### <a name="get-todos-parameters">[Parameters](#get-todos-parameters)</a>

| Name | Type | In | Description |
| ---- | ---- | -- | ----------- |
| N/A

### <a name="get-todos-response">[Response](#get-todos-response)</a>

```json
[
  {
    "id": "507f191e810c19729de860ea",
    "name": "Buy milk",
    "description": "Buy milk",
    "done": false,
    "repeating": "daily",
    "createdAt": "2016-01-01T00:00:00Z",
    "updatedAt": "2016-01-01T00:00:00Z"
  },
  {
    "id": "507f1f77bcf86cd799439011",
    "name": "Call Anna",
    "done": false,
    "createdAt": "2016-01-01T00:00:00Z",
    "updatedAt": "2016-01-01T00:00:00Z"
  },
  {
    "id": "54495ad94c934721ede76d90",
    "name": "Review notes",
    "description": "Check for grammatical errors",
    "done": true,
    "dueDate": "2016-01-02T00:00:00Z",
    "createdAt": "2016-01-01T00:00:00Z",
    "updatedAt": "2016-01-01T00:00:00Z"
  }
]
```

## [POST /todo](#post-todo)

Create a new todo item for the current user.

### <a name="post-todo-parameters">[Parameters](#post-todo-parameters)</a>

| Name        | Type    | In    | Description                           |
| ----------  | ------- | ----- | ------------------------------------- |
| name        | string  | body  | The todo name.                        |
| done        | boolean | body  | Optional: If the todo is done or not. |
| description | string  | body  | Optional: The todo description.       |
| dueDate     | date    | body  | Optional: The todo's due date.        |
| repeating   | string  | body  | Optional: The todo is repeating. Enum of "daily", "weekly", "monthly", "yearly". |

### <a name="post-todo-response">[Response](#post-todo-response)</a>

Possible errors include 400 Bad Request if your todo does not include a name or if the dueDate is incorrectly formatted.

```json
{
  "id": "507f191e810c19729de860ea",
  "name": "Buy milk",
  "description": "Buy milk",
  "done": false,
  "repeating": "daily",
  "createdAt": "2016-01-01T00:00:00Z",
  "updatedAt": "2016-01-01T00:00:00Z"
}
```

## [GET /todo/:id](#get-todoid)

Get one of the user's todo items. [Read the parameters for POST /todo to help understand the response of this endpoint fully.](#post-todo-parameters) `id`, `createdAt` and `updatedAt` are created by the server and cannot be edited directly.

### <a name="get-todo-id-parameters">[Parameters](#get-todo-id-parameters)</a>

| Name | Type    | In   | Description                     |
| ---- | ------- | ---- | ------------------------------- |
| id   | string  | path | The ID of the todo item to get. |

### <a name="get-todo-id-response">[Response](#get-todo-id-response)</a>

Possible errors include 404 Not Found if a todo with the given ID doesn't exist.

```json
{
  "id": "507f191e810c19729de860ea",
  "name": "Buy milk",
  "description": "Buy milk",
  "done": false,
  "repeating": "daily",
  "createdAt": "2016-01-01T00:00:00Z",
  "updatedAt": "2016-01-01T00:00:00Z"
}
```

## [PATCH /todo/:id](#patch-todoid)

Edit one of the user's todo items. [Read the parameters for POST /todo to help understand the response of this endpoint fully.](#post-todo-parameters) `id`, `createdAt` and `updatedAt` are created by the server and cannot be edited directly.

### <a name="patch-todo-id-parameters">[Parameters](#patch-todo-id-parameters)</a>

| Name        | Type    | In    | Description                                |
| ----------  | ------- | ----- | ------------------------------------------ |
| id          | string  | path  | The ID of the todo item to edit.           |
| name        | string  | body  | Optional: The todo name.                   |
| done        | boolean | body  | Optional: Whether the todo is done or not. |
| description | string  | body  | Optional: The todo description.            |
| repeating   | string  | body  | Optional: The todo is repeating. Enum of "daily", "weekly", "monthly", "yearly". |
| dueDate     | date    | body  | Optional: The todo's due date.             |

### <a name="patch-todo-id-response">[Response](#patch-todo-id-response)</a>

Possible errors include 404 Not Found if a todo with the given ID doesn't exist and 400 Bad Request if the dueDate is incorrectly formatted.

```json
{
  "id": "507f191e810c19729de860ea",
  "name": "Buy milk",
  "description": "Buy milk",
  "done": false,
  "repeating": "daily",
  "createdAt": "2016-01-01T00:00:00Z",
  "updatedAt": "2016-01-01T00:00:00Z"
}
```

Note: The endpoint returns the updated todo item.

## [DELETE /todo/:id](#delete-todoid)

Delete one of the user's todo items. [Read the parameters for POST /todo to help understand the response of this endpoint fully.](#post-todo-parameters) `id`, `createdAt` and `updatedAt` are created by the server and cannot be edited directly.

### <a name="delete-todo-id-parameters">[Parameters](#delete-todo-id-parameters)</a>

| Name | Type    | In   | Description                     |
| ---- | ------- | ---- | ------------------------------- |
| id   | string  | path | The ID of the todo item to get. |

### <a name="delete-todo-id-response">[Response](#delete-todo-id-response)</a>

Possible errors include 404 Not Found if a todo with the given ID doesn't exist.

```json
{
  "id": "507f191e810c19729de860ea",
  "name": "Buy milk",
  "description": "Buy milk",
  "done": false,
  "repeating": "daily",
  "createdAt": "2016-01-01T00:00:00Z",
  "updatedAt": "2016-01-01T00:00:00Z"
}
```
