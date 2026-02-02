# Handler Implementation Rules - Jimu Project

## 1. File Organization & Naming
* **Location**: `internal/handlers/`
* **Naming Convention**: `[domain]_handler.go` (e.g., `auth_handler.go`, `workout_handler.go`).
* **Test Naming**: `[domain]_handler_test.go`.
* **Mocks**: Place unit test mocks within the `_test.go` file to keep production binaries lean.

## 2. Dependency Injection & Interfaces
* **Rule**: Handlers must never depend on concrete repository structs.
* **Rule**: Define interfaces at the top of the handler file that describe the required behavior (e.g., `UserScanner`, `WorkoutScanner`).
* **Constructor**: Use a `New[Domain]Handler` function to inject these interfaces.

## 3. Handler Method Execution "Maze"
All handler functions must follow this exact sequence:

1.  **Context Check**: Retrieve `userID` from context via `middleware.UserIDKey`.
    * *If missing/invalid*: Return `401 Unauthorized`.
2.  **Request Decoding**: Use `json.NewDecoder(r.Body).Decode(&req)`.
    * *If failure*: Return `400 Bad Request`.
    **2.1 URL Parameter Standards (The Golden Rule)**
    - **Path parameters** (`/resource/{id}`): Use **only** to identify a single, specific resource for `GET`, `PUT`, or `DELETE`.
      - Example: `GET /workouts/123-abc`
    - **Query parameters** (`/resource?key=val`): Use **only** for `GET` requests that filter/search/paginate a collection.
      - Example: `GET /comments?workout_id=123-abc`

3.  **Repository Call**: Pass the context and decoded data to the injected interface.
4.  **Error Mapping**:
    * `repository.ErrProfileNotFound` $\rightarrow$ `404 Not Found`
    * `repository.ErrUsernameTaken` $\rightarrow$ `409 Conflict`
    * *Other errors*: Log the error and return `500 Internal Server Error`.
5.  **Response**: Set `Content-Type: application/json` and return the appropriate status (`200 OK`, `201 Created`, or `204 No Content`).

## 4. Unit Testing Standards
* **Table-Driven Tests**: Use a slice of structs to define test cases (Success, Unauthorized, Not Found, DB Failure).
* **Context Injection**: Use `testutils.InjectUserID(req, uid)` to simulate middleware.
* **Request Body**: Never send a `nil` body for `POST/PUT/PATCH` requests; use `strings.NewReader("{}")` at a minimum.
* **Mock Behavior**: Program mocks to return specific errors (e.g., `repository.ErrProfileNotFound`) to verify the handler's error mapping logic.

## 5. Implementation Template
```go
// internal/handlers/example_handler.go

type ExampleScanner interface {
    DoSomething(ctx context.Context, id uuid.UUID) error
}

type ExampleHandler struct {
    Repo ExampleScanner
}

func NewExampleHandler(r ExampleScanner) *ExampleHandler {
    return &ExampleHandler{Repo: r}
}

func (h *ExampleHandler) Handle(w http.ResponseWriter, r *http.Request) {
    // 1. Context Check
    // 2. Request Decoding
    // 3. Repository Call
    // 4. Error Mapping
    // 5. Response Construction
}