# Cerulean REST API Documentation

**This REST API is still a work in progress!** Some endpoints and features have not been finalised. If you are writing a Cerulean client, you will need to keep track of the current API until beta versions are released. Additionally, some of these endpoints have not yet been implemented in the back-end.

- Password length/security validation.
- Todo ordering endpoints.
- Multiple to-do lists.
- `POST /verifyuser`
- `POST /register`

All dates sent to and fro from the REST API are encoded as `ISO 8601` strings as emitted by Date#toISOString in JavaScript.

## [Authentication Scheme](#authentication-scheme)

For authentication, the Cerulean REST API requires an `Authorization` header or a `cerulean_token` cookie to be sent with every request, containing a token which is sent to the client after logging in using the [POST /login](#post-login) endpoint. The [POST /logout](#post-logout) endpoint can be used to invalidate the user's token. The token returned expires after 6 months. Your client should be well equipped to handle token expiries.

## [POST /login](#post-login)

Log into the Cerulean API and retrieve a token.

### [Parameters](#post-login-parameters)

| Name       | Type    | In    | Description                 |
| ---------- | ------- | ----- | --------------------------- |
| `username` | string  | body  | The username to login with. |
| `password` | string  | body  | The password to login with. |
| `cookie`   | boolean | query | Optional: Set to `false` to avoid getting `Set-Cookie: cerulean_token=` |

### [Response](#post-login-response)

```json
{
  "token": "JRPnrZPzeb8hi+RigUYZjIBWg4N1hImlI+AwKkfi4fk"
}
```

## [POST /logout](#post-logout)

Logout and invaliate the current token.

### [Parameters](#post-logout-parameters)

| Name | Type | In | Description |
| ---- | ---- | -- | ----------- |
| N/A

### [Response](#post-logout-response)

```json
{
  "success": true
}
```

## [POST /changepassword](#post-changepassword)

Change your current user's password. This also invalidates all tokens except your current one.

### [Parameters](#post-changepassword-parameters)

| Name              | Type   | In   | Description                       |
| ----------------- | ------ | ---- | --------------------------------- |
| `currentPassword` | string | body | The old password.                 |
| `newPassword`     | string | body | The new password you wish to set. |

### [Response](#post-changepassword-response)

```json
{
  "success": true
}
```

## [GET /todos](#get-todos)

Get all of the user's todo items. [Read the parameters for POST /todo to help understand the response of this endpoint fully.](#post-todo-parameters) `id`, `createdAt` and `updatedAt` are created by the server and cannot be edited directly.

### [Parameters](#get-todos-parameters)

| Name | Type | In | Description |
| ---- | ---- | -- | ----------- |
| N/A

### [Response](#get-todos-response)

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

Create a new to-do for the current user.

### [Parameters](#post-todo-parameters)

| Name        | Type    | In    | Description                      |
| ----------  | ------- | ----- | -------------------------------- |
| name        | string  | body  | The todo name.                   |
| done        | boolean | body  | Whether the todo is done or not. |
| description | string  | body  | Optional: The todo description.  |
| repeating   | string  | body  | Optional: The todo is repeating. Enum of "daily", "weekly", "monthly", "yearly". |
| dueDate     | date    | body  | Optional: The todo's due date.   |

### [Response](#post-todo-response)

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

## [GET /todo/:id](#get-todo-id)

Get one of the user's todo items. [Read the parameters for POST /todo to help understand the response of this endpoint fully.](#post-todo-parameters) `id`, `createdAt` and `updatedAt` are created by the server and cannot be edited directly.

### [Parameters](#get-todo-id-parameters)

| Name | Type    | In   | Description                     |
| ---- | ------- | ---- | ------------------------------- |
| id   | string  | path | The ID of the todo item to get. |

### [Response](#get-todo-id-response)

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

## [PATCH /todo/:id](#patch-todo-id)

Edit one of the user's todo items. [Read the parameters for POST /todo to help understand the response of this endpoint fully.](#post-todo-parameters) `id`, `createdAt` and `updatedAt` are created by the server and cannot be edited directly.

### [Parameters](#patch-todo-id-parameters)

| Name        | Type    | In    | Description                                |
| ----------  | ------- | ----- | ------------------------------------------ |
| id          | string  | path  | The ID of the todo item to edit.           |
| name        | string  | body  | Optional: The todo name.                   |
| done        | boolean | body  | Optional: Whether the todo is done or not. |
| description | string  | body  | Optional: The todo description.            |
| repeating   | string  | body  | Optional: The todo is repeating. Enum of "daily", "weekly", "monthly", "yearly". |
| dueDate     | date    | body  | Optional: The todo's due date.             |

### [Response](#patch-todo-id-response)

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

## [DELETE /todo/:id](#delete-todo-id)

Delete one of the user's todo items. [Read the parameters for POST /todo to help understand the response of this endpoint fully.](#post-todo-parameters) `id`, `createdAt` and `updatedAt` are created by the server and cannot be edited directly.

### [Parameters](#delete-todo-id-parameters)

| Name | Type    | In   | Description                     |
| ---- | ------- | ---- | ------------------------------- |
| id   | string  | path | The ID of the todo item to get. |

### [Response](#delete-todo-id-response)

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
