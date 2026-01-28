# Repository Implementation Rules - Jimu Project

## 1. File Structure
- `[entity]_repo.go`: Contains the struct and method logic.
- `queries.go`: Contains all raw SQL strings as constants.
- `errors.go`: Package-level sentinel errors (e.g., `ErrNotFound`).
- `utils.go`: Common logic for building dynamic queries.

## 2. CRUD Requirements
- **Create**: Use `Upsert` where possible to ensure idempotency. If a table is populated via DB Trigger, DO NOT implement a manual Insert.
- **Update**: 
  - **Input**: Use a separate `Update[Entity]Request` struct with pointers (e.g., `*string`).
  - **Null Handling**:
    - If field is `nil`: Skip (don't update).
    - If field is "Zero Value" (`""`, `0`, `time.Time.IsZero()`): Update DB column to `NULL`.
  - **Guard Clause**: If no fields are provided for update, return `nil` immediately.
- **Delete**: Check `RowsAffected`. If `0`, return `ErrNotFound`.

## 3. Access Guards & Privacy
- **Owner-Only Guard**: For `user_devices` and `subscriptions`, all queries MUST include a filter for the current `user_id` to prevent cross-user data access.
- **Admin Exception**: If the requester is present in the `sys_admins` table, the Owner-Only guard may be bypassed.
- **Social Isolation**: 
  - **Blocked Users**: Queries fetching social content must exclude rows where a block relationship exists.
  - **Private Profiles**: Content from private users must only be visible if a confirmed 'following' relationship exists.

## 4. Time & Standards
- **Data Types**: All time-related fields in the database MUST use `TIMESTAMPTZ` for timezone-agnostic storage.
- **Validation**: Use Go's `time.IsZero()` to validate timestamps before database persistence.

## 5. Error Handling
- Never return raw DB errors to the caller.
- Convert PostgreSQL error codes via `pgconn.PgError`:
  - `23505` (Unique Violation) -> `ErrAlreadyExists`.
  - `23503` (Foreign Key Violation) -> `ErrReferenceViolation`.
  - RowsAffected == 0 -> `ErrNotFound`.