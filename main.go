package main

import (
	"os"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
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
		go monitorChain(chainConfig, config.Slack.WebhookURL, config.DataDir, interval, &wg)
	}

	wg.Wait()
}

func monitorChain(chainConfig ChainConfig, slackWebhookURL string, dataDir string, interval time.Duration, wg *sync.WaitGroup) {
	defer wg.Done()

	log.Infof("Starting monitor for chain: %s", chainConfig.ChainID)

	// Load persisted proposals on startup
	persisted, err := loadPersistedProposals(dataDir, chainConfig.ChainID)
	if err != nil {
		log.Warnf("[%s] Failed to load persisted proposals: %s, starting fresh", chainConfig.ChainID, err)
		persisted = make(map[string]Proposal)
	} else {
		log.Infof("[%s] Loaded %d persisted proposals", chainConfig.ChainID, len(persisted))
	}

	for {
		uncompleted, err := fetchUncompletedProposals(chainConfig.Endpoint, chainConfig.Validator)
		if err != nil {
			log.Errorf("[%s] Error fetching uncompleted proposals: %s", chainConfig.ChainID, err)
			time.Sleep(interval)
			continue
		}

		// Deep copy uncompleted into all
		all := make(map[string]Proposal, len(uncompleted))
		for k, v := range uncompleted {
			all[k] = v
		}
		
		failed := false 
		for proposalID, persistedProposal := range persisted {
			currentProposal, err := fetchProposalByID(chainConfig.Endpoint, proposalID, chainConfig.Validator)
			if err != nil {
				log.Warnf("[%s] Error fetching latest status of persisted proposal %s: %s", 
					chainConfig.ChainID, proposalID, err)
				failed = true
				break 
			}

			if persistedProposal.Status != currentProposal.Status {
				all[proposalID] = *currentProposal
			}
		}

		if failed {
			time.Sleep(interval)
			continue
		}

		result := compareProposals(all, persisted)
		if len(result.ProposalChanges) == 0 && len(result.VoteChanges) == 0 {
			continue
		}
		
		// Send vote change notifications
		for _, voteChange := range result.VoteChanges {
			err = postVoteChangeToSlack(chainConfig, voteChange, slackWebhookURL)
			if err != nil {
				log.Errorf("[%s] Error posting vote change to slack: %s", chainConfig.ChainID, err)
				failed = true
				break
			}
		}
		
		// Send proposal change notifications (new proposals, status changes)
		for _, change := range result.ProposalChanges {
			err = postProposalChangeToSlack(chainConfig, change, slackWebhookURL)
			if err != nil {
				log.Errorf("[%s] Error posting to slack: %s", chainConfig.ChainID, err)
				failed = true
				break  
			}
		}

		if failed {
			time.Sleep(interval)
			continue
		}
		
		err = savePersistedProposals(dataDir, chainConfig.ChainID, uncompleted)
		if err != nil {
			//TODO: maybe we should notify to slack about this error. 
			log.Errorf("[%s] Error saving persisted proposals: %s", chainConfig.ChainID, err)
			time.Sleep(interval)
			continue
		}

		// Update persisted map for next iteration
		persisted = uncompleted // keep only uncompleted proposals TODO shall we do deep copy?
		time.Sleep(interval)
	}
}
