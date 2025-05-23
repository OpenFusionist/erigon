package spectest

import (
	"fmt"
	"io/fs"
	"os"

	clparams2 "github.com/erigontech/erigon/cl/clparams"
	"github.com/erigontech/erigon/cl/cltypes"
	"github.com/erigontech/erigon/cl/phase1/core/state"
	"github.com/erigontech/erigon/cl/utils"

	"gopkg.in/yaml.v3"

	"github.com/erigontech/erigon-lib/types/ssz"
)

func ReadMeta(root fs.FS, name string, obj any) error {
	bts, err := fs.ReadFile(root, name)
	if err != nil {
		return fmt.Errorf("couldnt read meta: %w", err)
	}
	err = yaml.Unmarshal(bts, obj)
	if err != nil {
		return fmt.Errorf("couldnt parse meta: %w", err)
	}
	return nil
}

func ReadYml(root fs.FS, name string, obj any) error {
	bts, err := fs.ReadFile(root, name)
	if err != nil {
		return fmt.Errorf("couldnt read meta: %w", err)
	}
	err = yaml.Unmarshal(bts, obj)
	if err != nil {
		return fmt.Errorf("couldnt parse meta: %w", err)
	}
	return nil
}

func ReadSsz(root fs.FS, version clparams2.StateVersion, name string, obj ssz.Unmarshaler) error {
	bts, err := fs.ReadFile(root, name)
	if err != nil {
		return fmt.Errorf("couldnt read meta: %w", err)
	}
	return utils.DecodeSSZSnappy(obj, bts, int(version))
}

func ReadSszOld(root fs.FS, obj ssz.Unmarshaler, version clparams2.StateVersion, name string) error {
	return ReadSsz(root, version, name, obj)
}

func ReadBeaconState(root fs.FS, version clparams2.StateVersion, name string) (*state.CachingBeaconState, error) {
	sszSnappy, err := fs.ReadFile(root, name)
	if err != nil {
		return nil, err
	}
	config := clparams2.MainnetBeaconConfig
	testState := state.New(&config)
	if err := utils.DecodeSSZSnappy(testState, sszSnappy, int(version)); err != nil {
		return nil, err
	}
	return testState, nil
}

func ReadBlock(root fs.FS, version clparams2.StateVersion, index int) (*cltypes.SignedBeaconBlock, error) {
	var blockBytes []byte
	var err error
	blockBytes, err = fs.ReadFile(root, fmt.Sprintf("blocks_%d.ssz_snappy", index))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	blk := cltypes.NewSignedBeaconBlock(&clparams2.MainnetBeaconConfig, version)
	if err = utils.DecodeSSZSnappy(blk, blockBytes, int(version)); err != nil {
		return nil, err
	}

	return blk, nil
}

func ReadBlockByPath(root fs.FS, version clparams2.StateVersion, path string) (*cltypes.SignedBeaconBlock, error) {
	var blockBytes []byte
	var err error
	blockBytes, err = fs.ReadFile(root, path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	blk := cltypes.NewSignedBeaconBlock(&clparams2.MainnetBeaconConfig, version)
	if err = utils.DecodeSSZSnappy(blk, blockBytes, int(version)); err != nil {
		return nil, err
	}

	return blk, nil
}

func ReadAnchorBlock(root fs.FS, version clparams2.StateVersion, name string) (*cltypes.BeaconBlock, error) {
	var blockBytes []byte
	var err error
	blockBytes, err = fs.ReadFile(root, name)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	blk := cltypes.NewBeaconBlock(&clparams2.MainnetBeaconConfig, version)
	if err = utils.DecodeSSZSnappy(blk, blockBytes, int(version)); err != nil {
		return nil, err
	}

	return blk, nil
}

func ReadBlockSlot(root fs.FS, index int) (uint64, error) {
	var blockBytes []byte
	var err error
	blockBytes, err = fs.ReadFile(root, fmt.Sprintf("blocks_%d.ssz_snappy", index))
	if os.IsNotExist(err) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}

	blockBytes, err = utils.DecompressSnappy(blockBytes, false)
	if err != nil {
		return 0, err
	}
	return ssz.UnmarshalUint64SSZ(blockBytes[100:108]), nil
}
func ReadBlocks(root fs.FS, version clparams2.StateVersion) ([]*cltypes.SignedBeaconBlock, error) {
	i := 0
	blocks := []*cltypes.SignedBeaconBlock{}
	var err error
	for {
		var blockBytes []byte
		blockBytes, err = fs.ReadFile(root, fmt.Sprintf("blocks_%d.ssz_snappy", i))
		if err != nil {
			break
		}
		blk := cltypes.NewSignedBeaconBlock(&clparams2.MainnetBeaconConfig, version)
		if err = utils.DecodeSSZSnappy(blk, blockBytes, int(version)); err != nil {
			return nil, err
		}
		blocks = append(blocks, blk)
		i++
	}
	if os.IsNotExist(err) {
		err = nil
	}
	return blocks, err
}
