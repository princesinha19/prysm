package kv

import (
	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	ethpb "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
	"github.com/prysmaticlabs/go-ssz"
	"github.com/prysmaticlabs/prysm/beacon-chain/core/helpers"
)

// SaveAggregatedAttestation saves an aggregated attestation in cache.
func (p *AttCaches) SaveAggregatedAttestation(att *ethpb.Attestation) error {
	if !helpers.IsAggregated(att) {
		return errors.New("attestation is not aggregated")
	}

	r, err := ssz.HashTreeRoot(att)
	if err != nil {
		return errors.Wrap(err, "could not tree hash attestation")
	}

	// DefaultExpiration is set to what was given to New(). In this case
	// it's one epoch.
	p.aggregatedAtt.Set(string(r[:]), att, cache.DefaultExpiration)

	return nil
}

// SaveAggregatedAttestations saves a list of aggregated attestations in cache.
func (p *AttCaches) SaveAggregatedAttestations(atts []*ethpb.Attestation) error {
	for _, att := range atts {
		if err := p.SaveAggregatedAttestation(att); err != nil {
			return err
		}
	}
	return nil
}

// AggregatedAttestations returns the aggregated attestations in cache.
func (p *AttCaches) AggregatedAttestations() []*ethpb.Attestation {
	atts := make([]*ethpb.Attestation, 0, p.aggregatedAtt.ItemCount())
	for s, i := range p.aggregatedAtt.Items() {
		// Type assertion for the worst case. This shouldn't happen.
		att, ok := i.Object.(*ethpb.Attestation)
		if !ok {
			p.aggregatedAtt.Delete(s)
		}
		atts = append(atts, att)
	}

	return atts
}

// AggregatedAttestationsBySlotIndex returns the aggregated attestations in cache,
// filtered by committee index and slot.
func (p *AttCaches) AggregatedAttestationsBySlotIndex(slot uint64, committeeIndex uint64) []*ethpb.Attestation {
	atts := make([]*ethpb.Attestation, 0, p.aggregatedAtt.ItemCount())
	for s, i := range p.aggregatedAtt.Items() {

		// Type assertion for the worst case. This shouldn't happen.
		att, ok := i.Object.(*ethpb.Attestation)
		if !ok {
			p.aggregatedAtt.Delete(s)
		}

		if slot == att.Data.Slot && committeeIndex == att.Data.CommitteeIndex {
			atts = append(atts, att)
		}
	}

	return atts
}

// DeleteAggregatedAttestation deletes the aggregated attestations in cache.
func (p *AttCaches) DeleteAggregatedAttestation(att *ethpb.Attestation) error {
	if !helpers.IsAggregated(att) {
		return errors.New("attestation is not aggregated")
	}

	r, err := ssz.HashTreeRoot(att)
	if err != nil {
		return errors.Wrap(err, "could not tree hash attestation")
	}

	p.aggregatedAtt.Delete(string(r[:]))

	return nil
}

// HasAggregatedAttestation checks if the input attestations has already existed in cache.
func (p *AttCaches) HasAggregatedAttestation(att *ethpb.Attestation) (bool, error) {
	r, err := ssz.HashTreeRoot(att)
	if err != nil {
		return false, errors.Wrap(err, "could not tree hash attestation")
	}

	for k := range p.aggregatedAtt.Items() {
		if k == string(r[:]) {
			return true, nil
		}
	}

	for k := range p.blockAtt.Items() {
		if k == string(r[:]) {
			return true, nil
		}
	}

	return false, nil
}
