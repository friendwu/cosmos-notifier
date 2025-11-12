package main

import (
	"fmt"

	slack "github.com/ashwanthkumar/slack-go-webhook"
)

func postProposalChangeToSlack(chainConfig ChainConfig, change ProposalChange, slackWebookUrl string) error {
	proposal := change.Proposal
	
	attachment := slack.Attachment{}

	//attachment.AddField(slack.Field{Title: "Chain:", Value: chainConfig.ChainID})
	attachment.AddField(slack.Field{Title: "ID", Value: proposal.ID})
	
	if change.IsNew {
		attachment.AddField(slack.Field{Title: "Status", Value: proposal.Status})
	} else if change.NewStatus != "" {
		attachment.AddField(slack.Field{Title: "Status Changed", Value: fmt.Sprintf("%s → %s", change.OldStatus, change.NewStatus)})
	}

	attachment.AddField(slack.Field{Title: "VotingStartTime", Value: proposal.VotingStartTime})
	attachment.AddField(slack.Field{Title: "VotingEndTime", Value: proposal.VotingEndTime})
	attachment.AddField(slack.Field{Title: "Title", Value: proposal.Title})
	attachment.AddField(slack.Field{Title: "Summary", Value: proposal.Summary})
	attachment.AddField(slack.Field{Title: "Proposer", Value: proposal.Proposer})

	if chainConfig.ExplorerURL != "" {
		explorerUrl := chainConfig.ExplorerURL
		if len(explorerUrl) > 0 && explorerUrl[len(explorerUrl)-1] != '/' {
			explorerUrl += "/"
		}
		explorerUrl += proposal.ID
		attachment.AddAction(slack.Action{Type: "button", Text: "Open", Url: explorerUrl, Style: "primary"})
	}

	var text string
	if change.IsNew {
		text = fmt.Sprintf("*NEW PROPOSAL: %s*", chainConfig.ChainID)
	} else {
		text = fmt.Sprintf("*PROPOSAL UPDATE: %s*", chainConfig.ChainID)
	}

	payload := slack.Payload{
		Text:        text,
		Username:    "Cosmos Proposal Notifier",
		IconEmoji:   ":monkey_face:",
		Attachments: []slack.Attachment{attachment},
	}
	err := slack.Send(slackWebookUrl, "", payload)
	if len(err) > 0 {
		return err[0]
	}
	return nil
}

func postVoteChangeToSlack(chainConfig ChainConfig, voteChange VoteChange, slackWebookUrl string) error {
	proposal := voteChange.Proposal
	
	attachment := slack.Attachment{}

	attachment.AddField(slack.Field{Title: "Proposal ID", Value: proposal.ID})
	attachment.AddField(slack.Field{Title: "Title", Value: proposal.Title})
	
	oldVoteStr := "not voted"
	if voteChange.OldVote != nil {
		oldVoteStr = voteChange.OldVote.GetOption()
		if oldVoteStr == "" {
			oldVoteStr = "not voted"
		}
	}
	
	newVoteStr := "not voted"
	if voteChange.NewVote != nil {
		newVoteStr = voteChange.NewVote.GetOption()
		if newVoteStr == "" {
			newVoteStr = "not voted"
		}
	}
	
	attachment.AddField(slack.Field{Title: "Vote Changed", Value: fmt.Sprintf("%s → %s", oldVoteStr, newVoteStr)})
	
	if chainConfig.Validator != "" {
		attachment.AddField(slack.Field{Title: "Validator", Value: chainConfig.Validator})
	}

	if chainConfig.ExplorerURL != "" {
		explorerUrl := chainConfig.ExplorerURL
		if len(explorerUrl) > 0 && explorerUrl[len(explorerUrl)-1] != '/' {
			explorerUrl += "/"
		}
		explorerUrl += proposal.ID
		attachment.AddAction(slack.Action{Type: "button", Text: "Open", Url: explorerUrl, Style: "primary"})
	}

	payload := slack.Payload{
		Text:        fmt.Sprintf("*VALIDATOR VOTE UPDATE: %s*", chainConfig.ChainID),
		Username:    "Cosmos Proposal Notifier",
		IconEmoji:   ":ballot_box_with_check:",
		Attachments: []slack.Attachment{attachment},
	}
	err := slack.Send(slackWebookUrl, "", payload)
	if len(err) > 0 {
		return err[0]
	}
	return nil
}
