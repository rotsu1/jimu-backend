# Jimu Backend: Integration Testing Specification

**Version:** 1.0  
**Last Updated:** 2026-01-28  
**Status:** Active

## 1. Philosophy & Objective

The goal of Integration Testing in **jimu** is to verify that the "Wiring" of the application works correctly. While Unit Tests verify logic in isolation (using mocks), Integration Tests verify the entire lifecycle of a request across boundaries.

**We adhere to the "Real Dependencies" Principle:**
* **No Mocks for Database:** Tests must run against a real PostgreSQL instance (via Docker).
* **No Mocks for Internal Logic:** Handlers call real Repositories.
* **Isolation:** Each test must run in a pristine state or clean up after itself.

---

## 2. Testing Levels

We divide our integration testing into two distinct layers.

### Level 1: Data Access Layer (Repositories)
* **Scope:** `internal/repository`
* **Goal:** Verify SQL queries, Schema constraints, and Transaction logic.
* **Current Status:** ✅ Implemented.
* **Dependencies:** Dockerized PostgreSQL (`jimu_test_db`).

### Level 2: HTTP Layer (Handlers & Router)
* **Scope:** `internal/handlers`, `internal/routers`
* **Goal:** Verify HTTP Status codes, JSON serialization, Middleware (Auth), and End-to-End data flow.
* **Current Status:** 🚧 **To Be Implemented** (Currently using mocks).
* **Strategy:** Spin up the `JimuRouter` with **Real Repositories** connected to the Test DB.

---

## 3. Infrastructure & Setup

### 3.1. The Test Environment
All integration tests rely on a dedicated, ephemeral environment.

* **Engine:** Docker Compose (`docker-compose.test.yml`).
* **Database:** PostgreSQL 16 (Alpine).
* **Migrations:** Must be applied automatically before the test suite runs.

### 3.2. Configuration
Tests should auto-detect the environment variables or default to the local test configuration:

| Variable | Value | Description |
| :--- | :--- | :--- |
| `DATABASE_URL` | `postgres://postgres:postgres@localhost:5433/jimu_test?sslmode=disable` | Connects to the Docker container. |
| `JIMU_SECRET` | `test-secret-key-123` | Used to sign/verify JWTs during tests. |

---

## 4. Implementation Specification

### 4.1. The "Test Server" Helper
To support **Level 2** testing, we will create a reusable helper struct `TestServer` in `internal/testutil`. This replaces the need to manually wire up handlers in every test function.

**Blueprint:**
```go
type TestServer struct {
    Router *router.JimuRouter
    DB     *pgxpool.Pool
    Config TestConfig
}

// NewTestServer performs the "Wiring" logic similar to main.go
// but connects to the Test DB.
func NewTestServer(t *testing.T) *TestServer {
    // 1. Connect to Test DB
    pool := SetupTestDB(t) 
    
    // 2. Initialize REAL Repos
    userRepo := repository.NewUserRepository(pool)
    workoutRepo := repository.NewWorkoutRepo(pool)
    // ... all other repos ...

    // 3. Initialize REAL Handlers
    authHandler := handlers.NewAuthHandler(userRepo, ...)
    workoutHandler := handlers.NewWorkoutHandler(workoutRepo)
    // ... all other handlers ...

    // 4. Create Router
    r := &router.JimuRouter{
        AuthHandler:    authHandler,
        WorkoutHandler: workoutHandler,
        JWTSecret:      "test-secret-key-123",
    }

    return &TestServer{Router: r, DB: pool}
}
```

### 4.2. Authentication Strategy
Since we cannot login with Google during tests, we must bypass the "External" auth step but keep the "Internal" token verification active.

**Helper Function:** `CreateTestToken(userID string) string`

**Mechanism:** Uses the static `JIMU_SECRET` ("test-secret-key-123") to sign a valid JWT.

**Usage:**
```go
token := testutil.CreateTestToken(user.ID.String())
req.Header.Set("Authorization", "Bearer " + token)
```

---

## 5. Test Case Standards

### 5.1. File Location
Integration tests for handlers should live in the same package but clearly marked.

* **Option A:** `internal/handlers/integration_test.go` (if package is handlers).
* **Option B (Preferred):** `tests/integration/workout_test.go` (Separate package `integration_test` to prevent circular imports and enforce black-box testing).

### 5.2. The "3-Step" Pattern
Every HTTP Integration test must follow this structure:

**Arrange (Seed):**
* Create a user in the DB directly via Repo.
* Create necessary parent data (e.g., existing Workouts).

**Act (Request):**
* Construct an HTTP Request (`httptest.NewRequest`).
* Add Auth Headers.
* Serve it via `server.Router.ServeHTTP`.

**Assert (Verify):**
* Check HTTP Status Code.
* Check JSON Response Body.
* **Crucial:** Query the DB to verify the side-effect (e.g., "Did the row count increase?").

### 5.3. Example Specification (Create Workout)
```go
func TestIntegration_CreateWorkout(t *testing.T) {
    // 1. Setup
    srv := testutil.NewTestServer(t)
    defer srv.DB.Close()
    
    // 2. Seed User
    user := srv.SeedUser(t, "integration-user")
    token := testutil.CreateTestToken(user.ID)

    // 3. Request
    payload := `{"title": "Chest Day", "start_at": "..."}`
    req := httptest.NewRequest("POST", "/workouts", strings.NewReader(payload))
    req.Header.Set("Authorization", "Bearer " + token)
    rr := httptest.NewRecorder()

    // 4. Execution
    srv.Router.ServeHTTP(rr, req)

    // 5. Assertions
    assert.Equal(t, http.StatusOK, rr.Code)
    
    // DB Verification
    var count int
    srv.DB.QueryRow(ctx, "SELECT count(*) FROM workouts WHERE user_id=$1", user.ID).Scan(&count)
    assert.Equal(t, 1, count)
}
```

---

## 6. Execution & CI/CD

### 6.1. Running Locally
We use a Makefile target to orchestrate the environment.

```bash
# Spins up DB, runs migrations, runs tests, tears down DB
make test-integration
```

### 6.2. CI Pipeline (GitHub Actions)
The pipeline will:

* Check out code.
* Start postgres service container.
* Run `go run cmd/migrate/main.go` (targeting the CI DB).
* Run `go test ./... -v`.

---

## 7. Migration Guide (From Unit to Integration)
To transition your current project state:

* [ ] Create `internal/testutil/server.go`: Implement the `NewTestServer` wiring logic.

* [ ] Refactor Handler Tests:
  * Select one handler (e.g., `WorkoutHandler`).
  * Create a new file `tests/integration/workout_test.go`.
  * Write a test that hits the `/workouts` endpoint using `NewTestServer`.

* [ ] Verify: Ensure it passes against the `docker-compose.test.yml` DB.

* [ ] Expand: Systematically cover all critical "Write" paths (Create/Update/Delete).