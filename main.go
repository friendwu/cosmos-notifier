package main

import (
	"os"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetOutput(os.Stdout)
}

func main() {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config.yaml"
	}

	config, err := loadConfig(configPath)
	if err != nil {
		log.Panicf("Failed to load config: %v", err)
	}

	setLogLevelFromConfig(config.LogLevel)
	interval := parseInterval(config.FetchInterval)

	log.Infof("Starting monitoring for %d chain(s)", len(config.Chains))

	var wg sync.WaitGroup
	for _, chainConfig := range config.Chains {
		wg.Add(1)
		go monitorChain(chainConfig, config.Slack.WebhookURL, interval, &wg)
	}

	wg.Wait()
}

func monitorChain(chainConfig ChainConfig, slackWebhookURL string, interval time.Duration, wg *sync.WaitGroup) {
	defer wg.Done()

	var lastProposalID string
	//firstIteration := true 
	//FIXME
	firstIteration := false

	log.Infof("Starting monitor for chain: %s", chainConfig.ChainID)

	for {
		latestProposal, err := getLatestProposal(chainConfig.Endpoint)
		if err != nil {
			log.Errorf("[%s] Error fetching proposal: %s", chainConfig.ChainID, err)
			time.Sleep(interval)
			continue
		}

		// post to slack only if didn't posted before
		if !firstIteration && lastProposalID != latestProposal.ID {
			err = postToSlack(chainConfig, *latestProposal, slackWebhookURL)
			if err != nil {
				log.Errorf("[%s] Error posting to slack: %s", chainConfig.ChainID, err)
			} else {
				log.Infof("[%s] Posted new proposal %s to slack", chainConfig.ChainID, latestProposal.ID)
			}
		}

		lastProposalID = latestProposal.ID
		firstIteration = false

		time.Sleep(interval)
	}
}
