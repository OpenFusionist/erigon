// Copyright 2021 The go-ethereum Authors
// (original work)
// Copyright 2024 The Erigon Authors
// (modifications)
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

package gasprice_test

import (
	"context"
	"errors"
	"testing"

	"github.com/erigontech/erigon-lib/kv/kvcache"
	"github.com/erigontech/erigon-lib/log/v3"
	"github.com/erigontech/erigon/eth/gasprice"
	"github.com/erigontech/erigon/eth/gasprice/gaspricecfg"
	"github.com/erigontech/erigon/rpc"
	"github.com/erigontech/erigon/rpc/jsonrpc"
	"github.com/erigontech/erigon/rpc/rpccfg"
)

func TestFeeHistory(t *testing.T) {

	overMaxQuery := make([]float64, 101)
	for i := 0; i < 101; i++ {
		overMaxQuery[i] = float64(1)
	}

	var cases = []struct {
		pending             bool
		maxHeader, maxBlock int
		count               int
		last                rpc.BlockNumber
		percent             []float64
		expFirst            uint64
		expCount            int
		expErr              error
	}{
		{false, 0, 0, 10, 30, nil, 21, 10, nil},
		{false, 0, 0, 10, 30, []float64{0, 10}, 21, 10, nil},
		{false, 0, 0, 10, 30, []float64{20, 10}, 0, 0, gasprice.ErrInvalidPercentile},
		{false, 0, 0, 1000000000, 30, nil, 0, 31, nil},
		{false, 0, 0, 1000000000, rpc.LatestBlockNumber, nil, 0, 33, nil},
		{false, 0, 0, 10, 40, nil, 0, 0, gasprice.ErrRequestBeyondHead},
		//{true, 0, 0, 10, 40, nil, 0, 0, gasprice.ErrRequestBeyondHead},
		{false, 20, 2, 100, rpc.LatestBlockNumber, nil, 13, 20, nil},
		{false, 20, 2, 100, rpc.LatestBlockNumber, []float64{0, 10}, 31, 2, nil},
		{false, 20, 2, 100, 32, []float64{0, 10}, 31, 2, nil},
		{false, 0, 0, 1, rpc.PendingBlockNumber, nil, 0, 0, nil},
		{false, 0, 0, 2, rpc.PendingBlockNumber, nil, 32, 1, nil},
		{false, 0, 0, 10, 30, overMaxQuery, 0, 0, gasprice.ErrInvalidPercentile},
		//{true, 0, 0, 2, rpc.PendingBlockNumber, nil, 32, 2, nil},
		//{true, 0, 0, 2, rpc.PendingBlockNumber, []float64{0, 10}, 32, 2, nil},
	}
	for i, c := range cases {
		config := gaspricecfg.Config{
			MaxHeaderHistory: c.maxHeader,
			MaxBlockHistory:  c.maxBlock,
		}

		func() {
			m := newTestBackend(t) //, big.NewInt(16), c.pending)
			defer m.Close()

			baseApi := jsonrpc.NewBaseApi(nil, kvcache.NewDummy(), m.BlockReader, false, rpccfg.DefaultEvmCallTimeout, m.Engine, m.Dirs, nil)
			tx, _ := m.DB.BeginTemporalRo(m.Ctx)
			defer tx.Rollback()

			cache := jsonrpc.NewGasPriceCache()
			oracle := gasprice.NewOracle(jsonrpc.NewGasPriceOracleBackend(tx, baseApi), config, cache, log.New())

			first, reward, baseFee, ratio, blobBaseFee, blobBaseFeeRatio, err := oracle.FeeHistory(context.Background(), c.count, c.last, c.percent)

			expReward := c.expCount
			if len(c.percent) == 0 {
				expReward = 0
			}
			expBaseFee := c.expCount
			if expBaseFee != 0 {
				expBaseFee++
			}

			if first.Uint64() != c.expFirst {
				t.Fatalf("Test case %d: first block mismatch, want %d, got %d", i, c.expFirst, first)
			}
			if len(reward) != expReward {
				t.Fatalf("Test case %d: reward array length mismatch, want %d, got %d", i, expReward, len(reward))
			}
			if len(baseFee) != expBaseFee {
				t.Fatalf("Test case %d: baseFee array length mismatch, want %d, got %d", i, expBaseFee, len(baseFee))
			}
			if len(ratio) != c.expCount {
				t.Fatalf("Test case %d: gasUsedRatio array length mismatch, want %d, got %d", i, c.expCount, len(ratio))
			}
			if c.expCount != 0 && len(blobBaseFee) != c.expCount+1 {
				t.Fatalf("Test case %d: blobBaseFee array length mismatch, want %d, got %d", i, c.expCount+1, len(blobBaseFee))
			}
			if len(blobBaseFeeRatio) != c.expCount {
				t.Fatalf("Test case %d: blobBaseFeeRatio array length mismatch, want %d, got %d", i, c.expCount, len(blobBaseFeeRatio))
			}
			if err != c.expErr && !errors.Is(err, c.expErr) {
				t.Fatalf("Test case %d: error mismatch, want %v, got %v", i, c.expErr, err)
			}
		}()
	}
}
