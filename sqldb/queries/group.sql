-- name: GetGroupByID :one
SELECT * FROM groups WHERE id = ?;

-- name: GetGroupByName :one
SELECT * FROM groups WHERE name = ? LIMIT 1;

-- name: ListGroups :many
SELECT * FROM groups ORDER BY created_at ASC;

-- name: CreateGroup :one
INSERT INTO groups (name, is_system)
VALUES (?, 0)
RETURNING *;

-- name: CreateSystemGroup :one
INSERT INTO groups (name, is_system)
VALUES (?, 1)
RETURNING *;

-- name: UpdateGroup :one
UPDATE groups SET name = ? WHERE id = ? AND is_system = 0
RETURNING *;

-- name: DeleteGroup :exec
DELETE FROM groups WHERE id = ? AND is_system = 0;

-- name: GetGroupPolicies :many
SELECT p.* FROM policy p
JOIN group_policy gp ON gp.policy_id = p.id
WHERE gp.group_id = ?;

-- name: CreatePolicy :one
INSERT INTO policy (action) VALUES (?) RETURNING *;

-- name: GetAllPolicies :many
SELECT * FROM policy ORDER BY action ASC;

-- name: DeletePolicyByID :exec
DELETE FROM policy WHERE id = ?;

-- name: AddPolicyToGroup :exec
INSERT INTO group_policy (group_id, policy_id) VALUES (?, ?);

-- name: RemovePolicyFromGroup :exec
DELETE FROM group_policy WHERE group_id = ? AND policy_id = ?;

-- name: GetUserFinalPermissions :many
SELECT p.action FROM group_policy gp
JOIN policy p ON p.id = gp.policy_id
JOIN organization_members om ON om.group_id = gp.group_id
WHERE om.user_id = ? AND om.organization_id = ?

UNION

SELECT p.action FROM user_policy up
JOIN policy p ON p.id = up.policy_id
WHERE up.user_id = ? AND up.org_id = ? AND up.effect = 'GRANT'

EXCEPT

SELECT p.action FROM user_policy up
JOIN policy p ON p.id = up.policy_id
WHERE up.user_id = ? AND up.org_id = ? AND up.effect = 'DENY';
