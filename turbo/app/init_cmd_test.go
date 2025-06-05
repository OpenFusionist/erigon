// Copyright 2024 The Erigon Authors
// This file is part of Erigon.
//
// Erigon is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Erigon is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with Erigon. If not, see <http://www.gnu.org/licenses/>.

package app

import (
	"os"
	"testing"

	"github.com/erigontech/erigon-lib/log/v3"
	"github.com/erigontech/erigon-lib/types"
	"github.com/erigontech/erigon/cmd/utils"
)

func TestGenesisJSONDecode(t *testing.T) {
	// Test genesis.json decode for endurance devnet
	genesisPath := "/Users/dengdiliang/ddl/fusionist-dev/devnet-deployer/genesis-data/el-cl-genesis-data/custom_config_data/genesis.json"

	// Create a logger for the test
	logger := log.New()
	logger.Info("Starting genesis JSON decode test", "path", genesisPath)

	// Open and decode the genesis file (same as in initGenesis function)
	file, err := os.Open(genesisPath)
	if err != nil {
		t.Fatalf("Failed to read genesis file: %v", err)
	}
	defer file.Close()

	genesis := new(types.Genesis)
	if err := parseGenesisStreaming(file, genesis, logger); err != nil {
		utils.Fatalf("invalid genesis file: %v", err)
	}

	// Verify basic genesis properties
	logger.Info("Genesis decoded successfully",
		"chain_id", genesis.Config.ChainID,
		"alloc_count", len(genesis.Alloc))

	t.Logf("Successfully decoded genesis file with chain ID: %v", genesis.Config.ChainID)
}
