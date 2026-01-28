# Testing Standards - Jimu Project

## 1. Get by ID
- **Normal**: Validate all fields match the inserted data.
- **NotFound**: Expect `pgx.ErrNoRows` or `ErrNotFound`.
- **Null Safety**: Ensure DB `NULL` correctly maps to Go `nil` pointers.

## 2. Update
- **Normal**: Update multiple fields and verify.
- **NotFound**: Attempt update on random UUID; expect `ErrNotFound`.
- **Zero Value to NULL**: Test that sending `""` or `time.Time{}` sets the DB column to `NULL`.
- **No Change**: Send an empty update request; expect `nil` error and no change to `updated_at`.
- **Constraint**: Attempt to update a field (like username) to a value that already exists; expect unique violation error.

## 3. Delete
- **Normal**: Delete and verify via Get.
- **Cascade**: Verify child records (e.g., settings) are also deleted.
- **NotFound**: Delete a random UUID; expect `ErrNotFound`.

## 4. Create / Upsert
- **Idempotency**: Run the same Upsert twice; verify no duplicate rows and same ID.
- **Defaults**: Ensure DB default values are populated in the returned model.
- **Trigger**: Check if side-effect rows (auto-created by DB) exist.
- **Null Constraint**: Test that missing NOT NULL fields return a DB error.

## 5. Pointer Safety (Anti-Segfault)
- Always copy IDs to a local variable (`targetID := profile.ID`) before deleting, to avoid referencing fields of a potentially nil struct after an error.