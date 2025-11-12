package main

import (
	log "github.com/sirupsen/logrus"
)

type ProposalChange struct {
	Proposal  Proposal
	IsNew     bool
	OldStatus string
	NewStatus string
}

type VoteChange struct {
	Proposal Proposal
	OldVote  *Vote
	NewVote  *Vote
}

// compareVotes compares two votes and returns true if they differ
func compareVotes(oldVote, newVote *Vote) bool {
	if oldVote == nil && newVote == nil {
		return false
	}
	if oldVote == nil || newVote == nil {
		return true // One voted, other didn't
	}
	// Compare the single option (first one in the array)
	return oldVote.GetOption() != newVote.GetOption()
}

// ProposalChangesResult holds both proposal changes and vote changes
type ProposalChangesResult struct {
	ProposalChanges []ProposalChange
	VoteChanges     []VoteChange
}

// compareProposals compares current proposals with persisted ones and returns changes
func compareProposals(current map[string]Proposal, persisted map[string]Proposal) ProposalChangesResult {
	var proposalChanges []ProposalChange
	var voteChanges []VoteChange

	// Check for new proposals, state changes, and vote changes
	for proposalID, currentProposal := range current {
		persistedProposal, exists := persisted[proposalID]
		
		if !exists {
			// New proposal
			proposalChanges = append(proposalChanges, ProposalChange{
				Proposal:  currentProposal,
				IsNew:     true,
				NewStatus: currentProposal.Status,
			})
			// if voted, add vote change
			if currentProposal.ValidatorVote != nil {
				voteChanges = append(voteChanges, VoteChange{
					Proposal: currentProposal,
					OldVote:  nil,
					NewVote:  currentProposal.ValidatorVote,
				})
			}
			log.Debugf("New proposal detected: %s (status: %s)", proposalID, currentProposal.Status)
		} else {
			// Check for status change
			statusChanged := persistedProposal.Status != currentProposal.Status
			// Check for vote change
			voteChanged := compareVotes(persistedProposal.ValidatorVote, currentProposal.ValidatorVote)
			
			if statusChanged {
				proposalChanges = append(proposalChanges, ProposalChange{
					Proposal:  currentProposal,
					IsNew:     false,
					OldStatus: persistedProposal.Status,
					NewStatus: currentProposal.Status,
				})
				log.Debugf("Proposal state changed: %s (%s -> %s)", proposalID, persistedProposal.Status, currentProposal.Status)
			}
			
			if voteChanged {
				oldVoteStr := "not voted"
				if persistedProposal.ValidatorVote != nil {
					oldVoteStr = persistedProposal.ValidatorVote.GetOption()
					if oldVoteStr == "" {
						oldVoteStr = "not voted"
					}
				}
				newVoteStr := "not voted"
				if currentProposal.ValidatorVote != nil {
					newVoteStr = currentProposal.ValidatorVote.GetOption()
					if newVoteStr == "" {
						newVoteStr = "not voted"
					}
				}
				log.Debugf("Proposal vote changed: %s (%s -> %s)", proposalID, oldVoteStr, newVoteStr)
				
				voteChanges = append(voteChanges, VoteChange{
					Proposal: currentProposal,
					OldVote:  persistedProposal.ValidatorVote,
					NewVote:  currentProposal.ValidatorVote,
				})
			}
		}
	}

	return ProposalChangesResult{
		ProposalChanges: proposalChanges,
		VoteChanges:     voteChanges,
	}
}

func proposalsToMap(proposals []Proposal) map[string]Proposal {
	result := make(map[string]Proposal)
	for _, p := range proposals {
		result[p.ID] = p
	}
	return result
}

