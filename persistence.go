package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type PersistedProposals struct {
	ChainID   string              `json:"chain_id"`
	Proposals map[string]Proposal `json:"proposals"` // keyed by proposal ID
}

func getDataFilePath(dataDir string, chainID string) string {
	return filepath.Join(dataDir, fmt.Sprintf("%s.json", chainID))
}

func loadPersistedProposals(dataDir string, chainID string) (map[string]Proposal, error) {
	filePath := getDataFilePath(dataDir, chainID)
	
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist yet, return empty map
			return make(map[string]Proposal), nil
		}
		return nil, fmt.Errorf("failed to read persisted proposals: %w", err)
	}

	var persisted PersistedProposals
	if err := json.Unmarshal(data, &persisted); err != nil {
		return nil, fmt.Errorf("failed to parse persisted proposals: %w", err)
	}

	return persisted.Proposals, nil
}

func savePersistedProposals(dataDir string, chainID string, proposals map[string]Proposal) error {
	// Create data directory if it doesn't exist
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	filePath := getDataFilePath(dataDir, chainID)
	persisted := PersistedProposals{
		ChainID:   chainID,
		Proposals: proposals,
	}

	data, err := json.MarshalIndent(persisted, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal proposals: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write persisted proposals: %w", err)
	}

	return nil
}

