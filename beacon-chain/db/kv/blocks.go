package kv

import (
	"bytes"
	"context"

	"github.com/boltdb/bolt"
	"github.com/gogo/protobuf/proto"
	"github.com/prysmaticlabs/prysm/beacon-chain/db/filters"
	ethpb "github.com/prysmaticlabs/prysm/proto/eth/v1alpha1"
)

// Block retrival by root.
func (k *Store) Block(ctx context.Context, blockRoot [32]byte) (*ethpb.BeaconBlock, error) {
	att := &ethpb.BeaconBlock{}
	err := k.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(blocksBucket)
		c := bkt.Cursor()
		for k, v := c.Seek(blockRoot[:]); k != nil && bytes.Contains(k, blockRoot[:]); k, v = c.Next() {
			if v != nil {
				return proto.Unmarshal(v, att)
			}
		}
		return nil
	})
	return att, err
}

// HeadBlock returns the latest canonical block in eth2.
// TODO(#3164): Implement.
func (k *Store) HeadBlock(ctx context.Context) (*ethpb.BeaconBlock, error) {
	return nil, nil
}

// Blocks retrieves a list of beacon blocks by filter criteria.
// TODO(#3164): Implement.
func (k *Store) Blocks(ctx context.Context, f *filters.QueryFilter) ([]*ethpb.BeaconBlock, error) {
	return nil, nil
}

// BlockRoots retrieves a list of beacon block roots by filter criteria.
// TODO(#3164): Implement.
func (k *Store) BlockRoots(ctx context.Context, f *filters.QueryFilter) ([][]byte, error) {
	return nil, nil
}

// HasBlock checks if a block by root exists in the db.
func (k *Store) HasBlock(ctx context.Context, blockRoot [32]byte) bool {
	exists := false
	// #nosec G104. Always returns nil.
	k.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(blocksBucket)
		c := bkt.Cursor()
		for k, v := c.Seek(blockRoot[:]); k != nil && bytes.Contains(k, blockRoot[:]); k, v = c.Next() {
			if v != nil {
				exists = true
				return nil
			}
		}
		return nil
	})
	return exists
}

// DeleteBlock by block root.
func (k *Store) DeleteBlock(ctx context.Context, blockRoot [32]byte) error {
	return k.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(blocksBucket)
		c := bkt.Cursor()
		for k, v := c.Seek(blockRoot[:]); k != nil && bytes.Contains(k, blockRoot[:]); k, v = c.Next() {
			if v != nil {
				return bkt.Delete(k)
			}
		}
		return nil
	})
}

// SaveBlock to the db.
// TODO(#3164): Implement.
func (k *Store) SaveBlock(ctx context.Context, block *ethpb.BeaconBlock) error {
	return nil
}

// SaveBlocks via batch updates to the db.
// TODO(#3164): Implement.
func (k *Store) SaveBlocks(ctx context.Context, blocks []*ethpb.BeaconBlock) error {
	return nil
}

// SaveHeadBlockRoot to the db.
// TODO(#3164): Implement.
func (k *Store) SaveHeadBlockRoot(ctx context.Context, blockRoot [32]byte) error {
	return nil
}
