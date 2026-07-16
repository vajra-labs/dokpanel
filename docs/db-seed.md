# Database Seeding Documentation

This document describes the database seeding and synchronization mechanism for permissions, policies, and system groups in Goploy.

---

## 1. Overview

Goploy uses a **code-first policy definitions** model. Instead of managing permissions manually via raw SQL scripts or migrators, all valid permissions (policies) and system groups (`ADMIN` and `MEMBER`) are declared inside the Go codebase.

During application startup, the seeding engine automatically reconciles these declarations with the database state.

---

## 2. Seeding Lifecycle & Architecture

The database seeding process is fully integrated with Go's **Uber Fx** dependency injection container and follows a strict execution flow:

```
[Application Startup]
        │
        ▼
[db.providerPool] ──► Synchronously sets SQLite Pragmas & runs Migrations
        │
        ▼
[seeds.SeedGroup] ──► Runs during the fx.Invoke lifecycle step (Sequential)
        │
        ├──► 1. syncPolicies (Inserts new, deletes obsolete policies)
        └──► 2. syncSystemGroups (Creates system groups & syncs mappings)
```

By running migrations and seeding synchronously during the database provider initialization, Goploy ensures that no API requests or handler invokes run against an unmigrated or unseeded database.

---

## 3. Code-Defined Policies

The master list of all valid resources and actions is declared in `src/types/group.go` within `DefaultStatements`:

- **Format:** `resource:action` (e.g., `project:create`, `service:read`).
- **Source of Truth:** If a permission is not present in `DefaultStatements`, it is considered obsolete.

---

## 4. Seeding Pipeline

The seeding engine performs the synchronization in two primary stages:

### Stage 1: Policies Synchronization (`syncPolicies`)

1. **Calculate Desired State:** Builds a set of all valid policy keys from `DefaultStatements`.
2. **Retrieve Current State:** Fetches all existing policy records from the database `policy` table.
3. **Reconcile Additions:** Inserts any policy key defined in the code but missing from the database.
4. **Reconcile Deletions:** Deletes any policy key found in the database but no longer present in the code.
5. **Cascade Effect:** Deleting a policy key automatically triggers an `ON DELETE CASCADE` on `group_policy` and `user_policy` tables, ensuring orphan mappings are cleaned up without database constraint violations.

---

### Stage 2: System Groups Synchronization (`syncSystemGroups`)

1. Goploy registers two primary system groups:
   - **`ADMIN`** (mapped to `types.AdminStatements`)
   - **`MEMBER`** (mapped to `types.MemberStatements`)

2. For each system group, the engine:
   - Checks if the group exists in the database. If not, it is created.
   - Begins a database transaction (`db.BeginTx`).
   - Fetches the current database policies assigned to the group.
   - **Assigns New Permissions:** Inserts mappings into the `group_policy` table for policies defined in the code statements but missing in the database.
   - **Revokes Deprecated Permissions:** Deletes mappings from the `group_policy` table for policies that were removed from the code statements.
   - Commits the transaction if successful, or rolls back on any error.

---

## 5. File Structure

All seeding components are located within the `src/db/seeds` directory:

| File                                                                              | Purpose                                                                              |
| :-------------------------------------------------------------------------------- | :----------------------------------------------------------------------------------- |
| [policy.go](file:///Users/aashish/Developer/golang/goploy/src/db/seeds/policy.go) | Implements `syncPolicies` to synchronize standard application permissions.           |
| [group.go](file:///Users/aashish/Developer/golang/goploy/src/db/seeds/group.go)   | Coordinates the seeding process and implements system groups policy synchronization. |

---

## 6. Benefits & Maintenance

- **Zero Schema drift:** Modifying `types/group.go` automatically updates the database on the next app restart.
- **Refactor Safety:** Deleted permissions are immediately cleaned up globally from all users and groups.
- **Transactional Reliability:** If syncing a system group fails, the transaction rolls back, keeping the database in a consistent state.
