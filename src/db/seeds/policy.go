package seeds

import (
	"context"

	"goploy/src/db/repos"
	"goploy/src/types"

	"github.com/rs/zerolog/log"
)

// syncPolicies syncs DefaultStatements with the policy database table.
func syncPolicies(ctx context.Context, query *repos.Queries) map[string]int64 {
	validPolicies := make(map[string]struct{})
	for resource, actions := range types.DefaultStatements {
		for _, action := range actions {
			validPolicies[resource+":"+string(action)] = struct{}{}
		}
	}
	existing, err := query.GetAllPolicies(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("Seed: failed to fetch policies")
	}
	existingMap := make(map[string]int64)
	for _, p := range existing {
		existingMap[p.Action] = p.ID
	}
	newCount := 0
	policyIDs := make(map[string]int64)
	for key := range validPolicies {
		if id, ok := existingMap[key]; ok {
			policyIDs[key] = id
			continue
		}
		p, insertErr := query.CreatePolicy(ctx, key)
		if insertErr != nil {
			log.Fatal().
				Err(insertErr).
				Str("Policy", key).
				Msg("Seed: failed to insert policy")
		}
		policyIDs[key] = p.ID
		newCount++
	}
	if newCount > 0 {
		log.Info().Int("Count", newCount).Msg("Seed: added new policies")
	}
	removedCount := 0
	for action, id := range existingMap {
		if _, ok := validPolicies[action]; !ok {
			if err = query.DeletePolicyByID(ctx, id); err != nil {
				log.Fatal().
					Err(err).
					Str("Policy", action).
					Msg("Seed: failed to delete outdated policy")
			}
			removedCount++
		}
	}
	if removedCount > 0 {
		log.Info().
			Int("Count", removedCount).
			Msg("Seed: removed outdated policies")
	}
	return policyIDs
}
