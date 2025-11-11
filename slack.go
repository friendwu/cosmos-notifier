package main

import (
	"fmt"

	slack "github.com/ashwanthkumar/slack-go-webhook"
)

func postToSlack(chainConfig ChainConfig, proposal Proposal, slackWebookUrl string) error {

	attachment := slack.Attachment{}
	attachment.AddField(slack.Field{Title: "Chain:", Value: chainConfig.ChainID})
	attachment.AddField(slack.Field{Title: "ID", Value: proposal.ID})
	attachment.AddField(slack.Field{Title: "Title", Value: proposal.Title})
	attachment.AddField(slack.Field{Title: "Summary", Value: proposal.Summary})
	attachment.AddField(slack.Field{Title: "Proposer", Value: proposal.Proposer})
	attachment.AddField(slack.Field{Title: "Status", Value: proposal.Status})
	attachment.AddField(slack.Field{Title: "VotingStartTime", Value: proposal.VotingStartTime})
	attachment.AddField(slack.Field{Title: "VotingEndTime", Value: proposal.VotingEndTime})

	if chainConfig.ExplorerURL != "" {
		explorerUrl := chainConfig.ExplorerURL
		if len(explorerUrl) > 0 && explorerUrl[len(explorerUrl)-1] != '/' {
			explorerUrl += "/"
		}
		explorerUrl += proposal.ID
		attachment.AddAction(slack.Action{Type: "button", Text: "Open", Url: explorerUrl, Style: "primary"})
	}

	payload := slack.Payload{
		Text:        fmt.Sprintf("*NEW PROPOSAL: %s*", chainConfig.ChainID),
		Username:    "robot",
		IconEmoji:   ":monkey_face:",
		Attachments: []slack.Attachment{attachment},
	}
	err := slack.Send(slackWebookUrl, "", payload)
	if len(err) > 0 {
		return err[0]
	}
	return nil
}
