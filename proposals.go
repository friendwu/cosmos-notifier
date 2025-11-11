package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
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

func fetchUncompletedProposals(cosmosEndpoint string) (map[string]Proposal, error) {
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

func fetchProposalByID(cosmosEndpoint string, proposalID string) (*Proposal, error) {
	url := fmt.Sprintf("%s/cosmos/gov/v1/proposals/%s", cosmosEndpoint, proposalID)

	resp, err := http.Get(url)
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

	return &response.Proposal, nil
}

