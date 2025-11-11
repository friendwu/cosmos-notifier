package main

import (
	log "github.com/sirupsen/logrus"
)

type ProposalChange struct {
	Proposal   Proposal
	IsNew      bool
	OldStatus  string
	NewStatus  string
}

// compareProposals compares current proposals with persisted ones and returns changes
func compareProposals(current map[string]Proposal, persisted map[string]Proposal) []ProposalChange {
	var changes []ProposalChange

	// Check for new proposals and state changes
	for proposalID, currentProposal := range current {
		persistedProposal, exists := persisted[proposalID]
		
		if !exists {
			// New proposal
			changes = append(changes, ProposalChange{
				Proposal:  currentProposal,
				IsNew:     true,
				NewStatus: currentProposal.Status,
			})
			log.Infof("New proposal detected: %s (status: %s)", proposalID, currentProposal.Status)
		} else if persistedProposal.Status != currentProposal.Status {
			// State changed
			changes = append(changes, ProposalChange{
				Proposal:  currentProposal,
				IsNew:     false,
				OldStatus: persistedProposal.Status,
				NewStatus: currentProposal.Status,
			})
			log.Infof("Proposal state changed: %s (%s -> %s)", proposalID, persistedProposal.Status, currentProposal.Status)
		}
	}

	return changes
}

func proposalsToMap(proposals []Proposal) map[string]Proposal {
	result := make(map[string]Proposal)
	for _, p := range proposals {
		result[p.ID] = p
	}
	return result
}

