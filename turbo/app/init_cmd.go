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
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/erigontech/erigon-lib/chain"
	"github.com/erigontech/erigon-lib/common"
	"github.com/erigontech/erigon-lib/common/datadir"
	"github.com/erigontech/erigon-lib/common/math"
	"github.com/erigontech/erigon-lib/kv"
	"github.com/erigontech/erigon-lib/log/v3"
	"github.com/erigontech/erigon-lib/types"
	"github.com/erigontech/erigon/cmd/utils"
	"github.com/erigontech/erigon/core"
	"github.com/erigontech/erigon/eth/tracers"
	"github.com/erigontech/erigon/node"
	"github.com/erigontech/erigon/turbo/debug"

	"net/http"
	_ "net/http/pprof"
)

var initCommand = cli.Command{
	Action:    MigrateFlags(initGenesis),
	Name:      "init",
	Usage:     "Bootstrap and initialize a new genesis block",
	ArgsUsage: "<genesisPath>",
	Flags: []cli.Flag{
		&utils.DataDirFlag,
		&utils.ChainFlag,
	},
	//Category: "BLOCKCHAIN COMMANDS",
	Description: `
The init command initializes a new genesis block and definition for the network.
This is a destructive action and changes the network in which you will be
participating.

It expects the genesis file as argument.`,
}

type genesisRaw struct {
	Config     json.RawMessage `json:"config"`
	Nonce      string          `json:"nonce"`
	Timestamp  float64         `json:"timestamp"`
	ExtraData  string          `json:"extraData"`
	GasLimit   string          `json:"gasLimit"`
	Difficulty string          `json:"difficulty"`
	Mixhash    string          `json:"mixhash"`
	Coinbase   string          `json:"coinbase"`
	ParentHash string          `json:"parentHash"`
	Alloc      json.RawMessage `json:"alloc"`
}

type allocAccountRaw struct {
	Balance     string          `json:"balance"`
	Nonce       string          `json:"nonce"`
	Code        string          `json:"code"`
	Constructor string          `json:"constructor"`
	Storage     json.RawMessage `json:"storage"`
}

func parseGenesisWithRawMessage(data []byte, logger log.Logger) (*types.Genesis, error) {
	var raw genesisRaw
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("initial shallow unmarshal failed: %w", err)
	}

	genesis := &types.Genesis{
		Config: &chain.Config{},
		Alloc:  make(types.GenesisAlloc),
	}

	genesis.Nonce = math.MustParseUint64(raw.Nonce)
	genesis.Timestamp = uint64(raw.Timestamp)
	genesis.ExtraData = common.FromHex(raw.ExtraData)
	genesis.GasLimit = math.MustParseUint64(raw.GasLimit)
	genesis.Difficulty = math.MustParseBig256(raw.Difficulty)
	genesis.Mixhash = common.HexToHash(raw.Mixhash)
	genesis.Coinbase = common.HexToAddress(raw.Coinbase)
	genesis.ParentHash = common.HexToHash(raw.ParentHash)

	if len(raw.Config) > 0 && string(raw.Config) != "null" {
		if err := json.Unmarshal(raw.Config, genesis.Config); err != nil {
			return nil, fmt.Errorf("unmarshal config failed: %w", err)
		}
	}

	if len(raw.Alloc) > 0 && string(raw.Alloc) != "null" {
		var allocMap map[string]allocAccountRaw
		if err := json.Unmarshal(raw.Alloc, &allocMap); err != nil {
			return nil, fmt.Errorf("unmarshal alloc map failed: %w", err)
		}

		for addrStr, accRaw := range allocMap {
			addr := common.HexToAddress(addrStr)
			account := types.GenesisAccount{
				Balance:     math.MustParseBig256(accRaw.Balance),
				Nonce:       math.MustParseUint64(accRaw.Nonce),
				Code:        common.FromHex(accRaw.Code),
				Constructor: common.FromHex(accRaw.Constructor),
			}

			if len(accRaw.Storage) > 0 && string(accRaw.Storage) != "null" {
				var storageMap map[string]string
				if err := json.Unmarshal(accRaw.Storage, &storageMap); err != nil {
					return nil, fmt.Errorf("unmarshal storage for address %s failed: %w", addrStr, err)
				}

				if len(storageMap) > 0 {
					account.Storage = make(map[common.Hash]common.Hash, len(storageMap))
					for keyStr, valStr := range storageMap {
						account.Storage[common.HexToHash(keyStr)] = common.HexToHash(valStr)
					}
				}
			}
			genesis.Alloc[addr] = account
		}
	}

	return genesis, nil
}

// initGenesis will initialise the given JSON format genesis file and writes it as
// the zero'd block (i.e. genesis) or will fail hard if it can't succeed.
func initGenesis(cliCtx *cli.Context) error {

	var logger log.Logger
	var tracer *tracers.Tracer
	var err error
	if logger, tracer, _, _, err = debug.Setup(cliCtx, true /* rootLogger */); err != nil {
		return err
	}

	go func() {
		logger.Info("Starting pprof on :6060")
		http.ListenAndServe("localhost:6060", nil)
	}()
	// Make sure we have a valid genesis JSON
	genesisPath := cliCtx.Args().First()
	if len(genesisPath) == 0 {
		utils.Fatalf("Must supply path to genesis JSON file")
	}

	data, err := os.ReadFile(genesisPath)
	if err != nil {
		utils.Fatalf("Failed to read genesis file: %v", err)
	}

	// Use optimized parsing instead of standard json.Decode
	genesis, err := parseGenesisWithRawMessage(data, logger)
	if err != nil {
		utils.Fatalf("invalid genesis file: %v", err)
	}

	logger.Info("after parseGenesisStreaming,GC")
	runtime.GC()
	if allocFile, err := os.Create("initgenesis_alloc_final.prof"); err == nil {
		pprof.Lookup("allocs").WriteTo(allocFile, 0)
		allocFile.Close()
		logger.Info("Allocation profile saved", "stage", "final", "file", "initgenesis_alloc_final.prof")
	}
	// DEBUG: just test json decode to save time
	time.Sleep(5 * time.Minute)
	return nil

	// Open and initialise both full and light databases
	stack, err := MakeNodeWithDefaultConfig(cliCtx, logger)
	if err != nil {
		return err
	}
	defer stack.Close()

	chaindb, err := node.OpenDatabase(cliCtx.Context, stack.Config(), kv.ChainDB, "", false, logger)
	if err != nil {
		utils.Fatalf("Failed to open database: %v", err)
	}

	if tracer != nil {
		if tracer.Hooks != nil && tracer.Hooks.OnBlockchainInit != nil {
			tracer.Hooks.OnBlockchainInit(genesis.Config)
		}
	}
	_, hash, err := core.CommitGenesisBlock(chaindb, genesis, datadir.New(cliCtx.String(utils.DataDirFlag.Name)), logger)
	if err != nil {
		utils.Fatalf("Failed to write genesis block: %v", err)
	}
	chaindb.Close()
	logger.Info("Successfully wrote genesis state", "hash", hash.Hash())
	return nil
}
