package seeds

import (
	"context"
	"database/sql"

	"goploy/src/db/repos"
	"goploy/src/types"

	"github.com/rs/zerolog/log"
)

var defaultGroups = map[string]types.Statements{
	"ADMIN":  types.AdminStatements,
	"MEMBER": types.MemberStatements,
}

// SeedGroup is the main entrypoint to seed.
func SeedGroup(db *sql.DB, query *repos.Queries) {
	ctx := context.Background()
	policyIDs := syncPolicies(ctx, query)
	syncSystemGroups(ctx, db, query, policyIDs)
}

// syncSystemGroups creates missing system groups and handles transaction for syncing policies.
func syncSystemGroups(
	ctx context.Context,
	db *sql.DB,
	query *repos.Queries,
	policyIDs map[string]int64,
) {
	for groupName, groupStatements := range defaultGroups {
		group, groupErr := query.GetGroupByName(ctx, groupName)
		if groupErr != nil && groupErr != sql.ErrNoRows {
			log.Fatal().
				Err(groupErr).
				Str("Group", groupName).
				Msg("Seed: failed to check group")
		}
		tx, txErr := db.BeginTx(ctx, nil)
		if txErr != nil {
			log.Fatal().Err(txErr).Msg("Seed: failed to begin transaction")
		}
		qtx := query.WithTx(tx)
		if groupErr == sql.ErrNoRows {
			group, groupErr = qtx.CreateSystemGroup(ctx, groupName)
			if groupErr != nil {
				_ = tx.Rollback()
				log.Fatal().
					Err(groupErr).
					Str("Group", groupName).
					Msg("Seed: failed to create group")
			}
			log.Info().Str("Group", groupName).Msg("Seed: created system group")
		}
		if err := syncSingleGroupPolicies(ctx, qtx, group, groupStatements, policyIDs); err != nil {
			_ = tx.Rollback()
			log.Fatal().
				Err(err).
				Str("Group", groupName).
				Msg("Seed: failed to sync group policies")
		}
		if err := tx.Commit(); err != nil {
			log.Fatal().Err(err).Msg("Seed: failed to commit transaction")
		}
	}
}

// syncSingleGroupPolicies syncs policy mappings for a single group in the database.
func syncSingleGroupPolicies(
	ctx context.Context,
	qtx *repos.Queries,
	group repos.Group,
	groupStatements types.Statements,
	policyIDs map[string]int64,
) error {
	currentDBPolicies, err := qtx.GetGroupPolicies(ctx, group.ID)
	if err != nil {
		return err
	}
	dbPoliciesMap := make(map[string]int64)
	for _, p := range currentDBPolicies {
		dbPoliciesMap[p.Action] = p.ID
	}
	desiredPolicies := make(map[string]struct{})
	for resource, actions := range groupStatements {
		for _, action := range actions {
			desiredPolicies[resource+":"+string(action)] = struct{}{}
		}
	}
	addedCount := 0
	for key := range desiredPolicies {
		if _, exists := dbPoliciesMap[key]; !exists {
			policyID, ok := policyIDs[key]
			if !ok {
				continue
			}
			err = qtx.AddPolicyToGroup(ctx, repos.AddPolicyToGroupParams{
				GroupID:  group.ID,
				PolicyID: policyID,
			})
			if err != nil {
				return err
			}
			addedCount++
		}
	}
	removedCount := 0
	for key, policyID := range dbPoliciesMap {
		if _, exists := desiredPolicies[key]; !exists {
			err = qtx.RemovePolicyFromGroup(
				ctx,
				repos.RemovePolicyFromGroupParams{
					GroupID:  group.ID,
					PolicyID: policyID,
				},
			)
			if err != nil {
				return err
			}
			removedCount++
		}
	}
	if addedCount > 0 || removedCount > 0 {
		log.Info().
			Str("Group", group.Name).
			Int("Added", addedCount).
			Int("Removed", removedCount).
			Msg("Seed: synced system group policies")
	}
	return nil
}
