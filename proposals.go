package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Define struct to match JSON structure
type ProposalsData struct {
	Proposals  []Proposal `json:"proposals"`
	Pagination interface{} `json:"pagination"` // We don't need pagination details
}

type Proposal struct {
	ID               string      `json:"id"`
	Messages         interface{} `json:"messages"`
	Status           string      `json:"status"`
	FinalTallyResult interface{} `json:"final_tally_result"`
	SubmitTime       string      `json:"submit_time"`
	DepositEndTime   string      `json:"deposit_end_time"`
	TotalDeposit     interface{} `json:"total_deposit"`
	VotingStartTime  string      `json:"voting_start_time"`
	VotingEndTime    string      `json:"voting_end_time"`
	Metadata         string      `json:"metadata"`
	Title            string      `json:"title"`
	Summary          string      `json:"summary"`
	Proposer         string      `json:"proposer"`
	ValidatorVote    *Vote       `json:"validator_vote,omitempty"` // Validator's vote for this proposal
}

type Vote struct {
	ProposalID string     `json:"proposal_id"`
	Voter      string     `json:"voter"`
	Options    []VoteOption `json:"options"` // API returns as array, but we only use the first one
	Metadata   string     `json:"metadata"`
}

// GetOption returns the first (and only) vote option
func (v *Vote) GetOption() string {
	if v == nil || len(v.Options) == 0 {
		return ""
	}
	return v.Options[0].Option
}

type VoteOption struct {
	Option string `json:"option"` // VOTE_OPTION_YES, NO, ABSTAIN, NO_WITH_VETO
	Weight string `json:"weight"`
}

func fetchProposals(cosmosEndpoint string) (map[string]Proposal, error) {
	baseURL := cosmosEndpoint + "/cosmos/gov/v1/proposals"
	
	params := url.Values{}
	params.Add("pagination.limit", "50") // Fetch one page, enough for active proposals
	params.Add("pagination.count_total", "true")
	params.Add("pagination.reverse", "true")
	
	fullURL := baseURL + "?" + params.Encode()

	resp, err := http.Get(fullURL)
	if err != nil {
		return make(map[string]Proposal), err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return make(map[string]Proposal), err
	}

	var proposalsData ProposalsData
	err = json.Unmarshal(body, &proposalsData)
	if err != nil {
		return make(map[string]Proposal), err
	}

	return proposalsToMap(proposalsData.Proposals), nil
}

func fetchUncompletedProposals(cosmosEndpoint string, validatorAddress string) (map[string]Proposal, error) {
	proposals, err := fetchProposals(cosmosEndpoint)
	if err != nil {
		return make(map[string]Proposal), err
	}

	if len(proposals) == 0 {
		return nil, fmt.Errorf("no proposals found")
	}

	uncompleted := make(map[string]Proposal)
	for proposalID, proposal := range proposals {
		if !isProposalCompleted(proposal.Status) {
			// Fetch validator vote if validator address is configured
			if validatorAddress != "" {
				vote, err := fetchProposalVotes(cosmosEndpoint, proposalID, validatorAddress)
				if err != nil {
					// Log but don't fail - vote fetching is optional
				} else {
					proposal.ValidatorVote = vote
				}
			}
			uncompleted[proposalID] = proposal
		}
	}
	return uncompleted, nil
}

func isProposalCompleted(status string) bool {
	completedStatuses := []string{
		"PROPOSAL_STATUS_PASSED",
		"PROPOSAL_STATUS_REJECTED",
		"PROPOSAL_STATUS_FAILED",
	}
	for _, s := range completedStatuses {
		if status == s {
			return true
		}
	}
	return false
}

func fetchProposalVotes(cosmosEndpoint string, proposalID string, validatorAddress string) (*Vote, error) {
	url := fmt.Sprintf("%s/cosmos/gov/v1/proposals/%s/votes", cosmosEndpoint, proposalID)

	client := http.Client{
		Timeout: 60 * time.Second,
	}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var votesResponse struct {
		Votes []Vote `json:"votes"`
	}
	err = json.Unmarshal(body, &votesResponse)
	if err != nil {
		return nil, err
	}

	// Find the validator's vote
	for _, vote := range votesResponse.Votes {
		if vote.Voter == validatorAddress {
			return &vote, nil
		}
	}

	// Validator hasn't voted yet
	return nil, nil
}

func fetchProposalByID(cosmosEndpoint string, proposalID string, validatorAddress string) (*Proposal, error) {
	url := fmt.Sprintf("%s/cosmos/gov/v1/proposals/%s", cosmosEndpoint, proposalID)

	client := http.Client{
		Timeout: 60 * time.Second,
	}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// The API returns { "proposal": {...} }
	var response struct {
		Proposal Proposal `json:"proposal"`
	}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	// Fetch validator vote if validator address is configured
	if validatorAddress != "" {
		vote, err := fetchProposalVotes(cosmosEndpoint, proposalID, validatorAddress)
		if err != nil {
			return nil, fmt.Errorf("error fetching validator vote for proposal %s: %s", proposalID, err)
		} else {
			response.Proposal.ValidatorVote = vote
		}
	}

	return &response.Proposal, nil
}

