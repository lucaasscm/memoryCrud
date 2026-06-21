# memoryCrud

A small in-memory user CRUD REST API written in Go using only the standard
library `net/http` router (Go 1.22+ method/path patterns), plus
[`google/uuid`](https://github.com/google/uuid) for IDs. Built as a study
project from Rocketseat's Go course.

Users are stored in a `map` in memory, so data is reset every time the server
restarts.

## Requirements

- Go 1.22 or newer

## Running

```bash
go run .
```

The server listens on `http://localhost:8080`.

## Endpoints

| Method   | Path              | Description          |
| -------- | ----------------- | -------------------- |
| `GET`    | `/api/users`      | List all users       |
| `GET`    | `/api/users/{id}` | Get a user by ID     |
| `POST`   | `/api/users`      | Create a user        |
| `PUT`    | `/api/users/{id}` | Update a user by ID  |
| `DELETE` | `/api/users/{id}` | Delete a user by ID  |

### User payload

```json
{
  "first_name": "Jane",
  "last_name": "Doe",
  "biography": "Software developer"
}
```

### Example

```bash
# Create a user
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"first_name":"Jane","last_name":"Doe","biography":"Software developer"}'

# List users
curl http://localhost:8080/api/users
```
