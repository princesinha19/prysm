package blockchain

import (
	"bytes"
	"context"
	"fmt"

	"github.com/pkg/errors"
	ethpb "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
	"github.com/prysmaticlabs/prysm/beacon-chain/core/feed"
	"github.com/prysmaticlabs/prysm/shared/bytesutil"
	"github.com/prysmaticlabs/prysm/shared/params"
	"github.com/prysmaticlabs/prysm/shared/slotutil"
	"github.com/sirupsen/logrus"
	"go.opencensus.io/trace"
)

// AttestationReceiver interface defines the methods of chain service receive and processing new attestations.
type AttestationReceiver interface {
	ReceiveAttestationNoPubsub(ctx context.Context, att *ethpb.Attestation) error
}

// ReceiveAttestationNoPubsub is a function that defines the operations that are preformed on
// attestation that is received from regular sync. The operations consist of:
//  1. Validate attestation, update validator's latest vote
//  2. Apply fork choice to the processed attestation
//  3. Save latest head info
func (s *Service) ReceiveAttestationNoPubsub(ctx context.Context, att *ethpb.Attestation) error {
	ctx, span := trace.StartSpan(ctx, "beacon-chain.blockchain.ReceiveAttestationNoPubsub")
	defer span.End()

	// Update forkchoice store for the new attestation
	if err := s.forkChoiceStore.OnAttestation(ctx, att); err != nil {
		return errors.Wrap(err, "could not process attestation from fork choice service")
	}

	// Run fork choice for head block after updating fork choice store.
	headRoot, err := s.forkChoiceStore.Head(ctx)
	if err != nil {
		return errors.Wrap(err, "could not get head from fork choice service")
	}
	// Only save head if it's different than the current head.
	if !bytes.Equal(headRoot, s.HeadRoot()) {
		signed, err := s.beaconDB.Block(ctx, bytesutil.ToBytes32(headRoot))
		if err != nil {
			return errors.Wrap(err, "could not compute state from block head")
		}
		if signed == nil || signed.Block == nil {
			return errors.New("nil head block")
		}
		if err := s.saveHead(ctx, signed, bytesutil.ToBytes32(headRoot)); err != nil {
			return errors.Wrap(err, "could not save head")
		}
	}

	processedAttNoPubsub.Inc()
	return nil
}

// This processes attestations from the attestation pool to account for validator votes and fork choice.
func (s *Service) processAttestation() {
	// Wait for state to be initialized.
	stateChannel := make(chan *feed.Event, 1)
	stateSub := s.stateNotifier.StateFeed().Subscribe(stateChannel)
	<-stateChannel
	stateSub.Unsubscribe()

	st := slotutil.GetSlotTicker(s.genesisTime, params.BeaconConfig().SecondsPerSlot)
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-st.C():
			ctx := context.Background()
			atts := s.attPool.ForkchoiceAttestations()
			for _, a := range atts {
				hasState := s.beaconDB.HasState(ctx, bytesutil.ToBytes32(a.Data.BeaconBlockRoot))
				hasBlock := s.beaconDB.HasBlock(ctx, bytesutil.ToBytes32(a.Data.BeaconBlockRoot))
				if !(hasState && hasBlock) {
					continue
				}

				if err := s.attPool.DeleteForkchoiceAttestation(a); err != nil {
					log.WithError(err).Error("Could not delete fork choice attestation in pool")
				}

				if err := s.ReceiveAttestationNoPubsub(ctx, a); err != nil {
					log.WithFields(logrus.Fields{
						"targetRoot": fmt.Sprintf("%#x", a.Data.Target.Root),
					}).WithError(err).Error("Could not receive attestation in chain service")
				}
			}
		}
	}
}
