# Cerulean REST API Documentation

**This REST API is still a work in progress!** Some endpoints and features have not been finalised. If you are writing a Cerulean client, you will need to keep track of the current API until beta versions are released. Additionally, some of these endpoints have not yet been implemented in the back-end.

- Multiple to-do lists.
- `POST /register`
- `POST /verifyuser`
- `POST /changepassword`
- `DELETE /todo/:id`
- `PATCH /todo/:id`
- `GET /todo/:id`
- `GET /todos`

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

| Name       | Type    | In    | Description                 |
| ---------- | ------- | ----- | --------------------------- |
| N/A

### [Response](#post-logout-response)

```json
{
  "success": true
}
```

## [POST /todo](#post-todo)

Create a new to-do for the current user.

### [Parameters](#post-todo-parameters)

| Name        | Type    | In    | Description                     |
| ----------  | ------- | ----- | ------------------------------- |
| name        | string  | body  | The todo name.                  |
| description | string  | body  | Optional: The todo description. |
| done        | boolean | body  | Optional: The todo is done.     |
| repeating   | string  | body  | Optional: The todo is repeating. Enum of "daily", "weekly", "monthly", "yearly". |
| dueDate     | date    | body  | Optional: The todo's due date.  |

### [Response](#post-todo-response)

```json
{
  "id": "507f191e810c19729de860ea",
  "name": "Buy milk",
  "description": "Buy milk",
  "done": false,
  "repeating": "daily",
  "dueDate": "2016-01-01T00:00:00Z",
  "createdAt": "2016-01-01T00:00:00Z",
  "updatedAt": "2016-01-01T00:00:00Z"
}
```
