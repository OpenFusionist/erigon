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

package jsonrpc

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"
	"testing"

	"github.com/erigontech/erigon-lib/chain"
	"github.com/erigontech/erigon-lib/common/datadir"
	"github.com/erigontech/erigon-lib/common/hexutil"
	"github.com/erigontech/erigon-lib/crypto"
	"github.com/erigontech/erigon-lib/kv/kvcache"
	"github.com/erigontech/erigon-lib/log/v3"
	"github.com/erigontech/erigon-lib/types"
	"github.com/erigontech/erigon/eth/ethconfig"
	"github.com/erigontech/erigon/execution/abi/bind"
	"github.com/erigontech/erigon/execution/abi/bind/backends"
	"github.com/erigontech/erigon/rpc"
	"github.com/erigontech/erigon/rpc/ethapi"
	"github.com/erigontech/erigon/rpc/jsonrpc/contracts"
	"github.com/erigontech/erigon/rpc/rpccfg"
)

// block 1 contains 3 Transactions
//	 1. deploy token A
//	 2. mint address 2 100 token
//	 3. transfer from address 2 to address 1

// test 2 bundles
// check balance of addr1 and addr 2 at the end of block and interblock

func TestCallMany(t *testing.T) {
	var (
		key, _   = crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
		key1, _  = crypto.HexToECDSA("49a7b37aa6f6645917e7b807e9d1c00d4fa71f18343b0d4122a4d2df64dd6fee")
		key2, _  = crypto.HexToECDSA("8a1f9a8f95be41cd7ccb6168179afb4504aefe388d1e14474d32c45c72ce7b7a")
		address  = crypto.PubkeyToAddress(key.PublicKey)
		address1 = crypto.PubkeyToAddress(key1.PublicKey)
		address2 = crypto.PubkeyToAddress(key2.PublicKey)
		gspec    = &types.Genesis{
			Config: chain.TestChainConfig,
			Alloc: types.GenesisAlloc{
				address:  {Balance: big.NewInt(9000000000000000000)},
				address1: {Balance: big.NewInt(200000000000000000)},
				address2: {Balance: big.NewInt(300000000000000000)},
			},
			GasLimit: 10000000,
		}
		chainID = big.NewInt(1337)
		ctx     = context.Background()

		addr1BalanceCheck = "70a08231" + "000000000000000000000000" + address1.Hex()[2:]
		addr2BalanceCheck = "70a08231" + "000000000000000000000000" + address2.Hex()[2:]
		transferAddr2     = "70a08231" + "000000000000000000000000" + address1.Hex()[2:] + "0000000000000000000000000000000000000000000000000000000000000064"
	)

	hexBytes, _ := hex.DecodeString(addr2BalanceCheck)
	balanceCallAddr2 := hexutil.Bytes(hexBytes)
	hexBytes, _ = hex.DecodeString(addr1BalanceCheck)
	balanceCallAddr1 := hexutil.Bytes(hexBytes)
	hexBytes, _ = hex.DecodeString(transferAddr2)
	transferCallData := hexutil.Bytes(hexBytes)

	//submit 3 Transactions and commit the results
	transactOpts, _ := bind.NewKeyedTransactorWithChainID(key, chainID)
	transactOpts1, _ := bind.NewKeyedTransactorWithChainID(key1, chainID)
	transactOpts2, _ := bind.NewKeyedTransactorWithChainID(key2, chainID)
	contractBackend := backends.NewTestSimulatedBackendWithConfig(t, gspec.Alloc, gspec.Config, gspec.GasLimit)
	defer contractBackend.Close()
	stateCache := kvcache.New(kvcache.DefaultCoherentConfig)
	tokenAddr, _, tokenContract, _ := contracts.DeployToken(transactOpts, contractBackend, address1)
	tokenContract.Mint(transactOpts1, address2, big.NewInt(100))
	tokenContract.Transfer(transactOpts2, address1, big.NewInt(100))
	contractBackend.Commit()

	// set up the callargs
	var nonce hexutil.Uint64 = 1
	var secondNonce hexutil.Uint64 = 2

	db := contractBackend.DB()
	engine := contractBackend.Engine()
	api := NewEthAPI(NewBaseApi(nil, stateCache, contractBackend.BlockReader(), false, rpccfg.DefaultEvmCallTimeout, engine, datadir.New(t.TempDir()), nil), db, nil, nil, nil, 5000000, ethconfig.Defaults.RPCTxFeeCap, 100_000, false, 100_000, 128, log.New())

	callArgAddr1 := ethapi.CallArgs{From: &address, To: &tokenAddr, Nonce: &nonce,
		MaxPriorityFeePerGas: (*hexutil.Big)(big.NewInt(1e9)),
		MaxFeePerGas:         (*hexutil.Big)(big.NewInt(1e10)),
		Data:                 &balanceCallAddr1,
	}
	callArgAddr2 := ethapi.CallArgs{From: &address, To: &tokenAddr, Nonce: &secondNonce,
		MaxPriorityFeePerGas: (*hexutil.Big)(big.NewInt(1e9)),
		MaxFeePerGas:         (*hexutil.Big)(big.NewInt(1e10)),
		Data:                 &balanceCallAddr2,
	}

	callArgTransferAddr2 := ethapi.CallArgs{From: &address2, To: &tokenAddr, Nonce: &nonce,
		MaxPriorityFeePerGas: (*hexutil.Big)(big.NewInt(1e9)),
		MaxFeePerGas:         (*hexutil.Big)(big.NewInt(1e10)),
		Data:                 &transferCallData,
	}

	timeout := int64(50000)
	txIndex := -1
	res, err := api.CallMany(ctx, []Bundle{{
		Transactions: []ethapi.CallArgs{callArgAddr1, callArgAddr2}}}, StateContext{BlockNumber: rpc.BlockNumberOrHashWithNumber(rpc.LatestBlockNumber), TransactionIndex: &txIndex}, nil, &timeout)
	if err != nil {
		t.Errorf("eth_callMany: %v", err)
	}

	// parse the results and do balance checks
	addr1CalRet := fmt.Sprintf("%v", res[0][0]["value"])[2:]
	addr2CalRet := fmt.Sprintf("%v", res[0][1]["value"])[2:]
	addr1Balance, err := strconv.ParseInt(addr1CalRet, 16, 64)
	if err != nil {
		t.Errorf("eth_callMany: %v", err)
	}
	addr2Balance, err := strconv.ParseInt(addr2CalRet, 16, 64)

	if err != nil {
		t.Errorf("eth_callMany: %v", err)
	}
	if addr1Balance != 100 || addr2Balance != 0 {
		t.Errorf("eth_callMany: %v", "balanceUnmatch")
	}

	txIndex = 2
	res, err = api.CallMany(ctx, []Bundle{{
		Transactions: []ethapi.CallArgs{callArgAddr1, callArgAddr2}}}, StateContext{BlockNumber: rpc.BlockNumberOrHashWithNumber(1), TransactionIndex: &txIndex}, nil, &timeout)
	if err != nil {
		t.Errorf("eth_callMany: %v", err)
	}

	addr1CalRet = fmt.Sprintf("%v", res[0][0]["value"])[2:]
	addr2CalRet = fmt.Sprintf("%v", res[0][1]["value"])[2:]
	addr1Balance, err = strconv.ParseInt(addr1CalRet, 16, 64)
	if err != nil {
		t.Errorf("%v", err)
	}
	addr2Balance, err = strconv.ParseInt(addr2CalRet, 16, 64)
	if err != nil {
		t.Errorf("%v", err)
	}

	if addr1Balance != 0 || addr2Balance != 100 {
		t.Errorf("eth_callMany: %s", "balanceUnmatch")
	}
	txIndex = -1
	res, err = api.CallMany(ctx, []Bundle{{Transactions: []ethapi.CallArgs{callArgTransferAddr2, callArgAddr1, callArgAddr2}}}, StateContext{BlockNumber: rpc.BlockNumberOrHashWithNumber(rpc.LatestBlockNumber), TransactionIndex: &txIndex}, nil, &timeout)
	if err != nil {
		t.Errorf("%v", err)
	}

	addr1CalRet = fmt.Sprintf("%v", res[0][1]["value"])[2:]
	addr2CalRet = fmt.Sprintf("%v", res[0][2]["value"])[2:]

	addr1Balance, err = strconv.ParseInt(addr1CalRet, 16, 64)
	if err != nil {
		t.Errorf("%v", err)
	}
	addr2Balance, err = strconv.ParseInt(addr2CalRet, 16, 64)
	if err != nil {
		t.Errorf("%v", err)
	}
	if addr1Balance != 100 || addr2Balance != 0 {
		t.Errorf("eth_callMany: %s", "balanceUnmatch")
	}
}
