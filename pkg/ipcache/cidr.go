// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Cilium

package ipcache

import (
	"context"
	"net/netip"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/cilium/cilium/pkg/identity"
	"github.com/cilium/cilium/pkg/labels"
	"github.com/cilium/cilium/pkg/logging/logfields"
	"github.com/cilium/cilium/pkg/metrics"
	"github.com/cilium/cilium/pkg/option"
	"github.com/cilium/cilium/pkg/source"
)

// AllocateCIDRs attempts to allocate identities for a list of CIDRs. If any
// allocation fails, all allocations are rolled back and the error is returned.
// When an identity is freshly allocated for a CIDR, it is added to the
// ipcache if 'newlyAllocatedIdentities' is 'nil', otherwise the newly allocated
// identities are placed in 'newlyAllocatedIdentities' and it is the caller's
// responsibility to upsert them into ipcache by calling upsertGeneratedIdentities().
//
// Upon success, the caller must also arrange for the resulting identities to
// be released via a subsequent call to ReleaseCIDRIdentitiesByCIDR().
//
// Deprecated: Prefer UpsertLabels() instead.
func (ipc *IPCache) AllocateCIDRs(
	prefixes []netip.Prefix, newlyAllocatedIdentities map[netip.Prefix]*identity.Identity,
) ([]*identity.Identity, error) {
	// maintain list of used identities to undo on error
	usedIdentities := make([]*identity.Identity, 0, len(prefixes))

	// Maintain list of newly allocated identities to update ipcache,
	// but upsert them to ipcache only if no map was given by the caller.
	upsert := false
	if newlyAllocatedIdentities == nil {
		upsert = true
		newlyAllocatedIdentities = map[netip.Prefix]*identity.Identity{}
	}

	allocateCtx, cancel := context.WithTimeout(context.Background(), option.Config.IPAllocationTimeout)
	defer cancel()

	ipc.metadata.RLock()
	ipc.Lock()
	allocatedIdentities := make(map[netip.Prefix]*identity.Identity, len(prefixes))
	for _, prefix := range prefixes {
		info := ipc.metadata.getLocked(prefix)

		oldNID := info.RequestedIdentity().ID()
		id, isNew, err := ipc.resolveIdentity(allocateCtx, prefix, info, oldNID)
		if err != nil {
			ipc.IdentityAllocator.ReleaseSlice(context.Background(), usedIdentities)
			ipc.Unlock()
			ipc.metadata.RUnlock()
			return nil, err
		}

		usedIdentities = append(usedIdentities, id)
		allocatedIdentities[prefix] = id
		if isNew {
			newlyAllocatedIdentities[prefix] = id
		}
	}
	ipc.Unlock()
	ipc.metadata.RUnlock()

	// Insert any newly allocated identities in to the policy engine
	addedIdentities := make(map[identity.NumericIdentity]labels.LabelArray, len(newlyAllocatedIdentities))
	for _, id := range newlyAllocatedIdentities {
		addedIdentities[id.ID] = id.LabelArray
	}
	ipc.UpdatePolicyMaps(context.TODO(), addedIdentities, nil)

	// Only upsert into ipcache if identity wasn't allocated
	// before and the caller does not care doing this
	if upsert {
		ipc.upsertGeneratedIdentities(newlyAllocatedIdentities, usedIdentities)
	}

	identities := make([]*identity.Identity, 0, len(allocatedIdentities))
	for _, id := range allocatedIdentities {
		identities = append(identities, id)
	}
	return identities, nil
}

func cidrLabelToPrefix(id *identity.Identity) (prefix netip.Prefix, ok bool) {
	var err error

	label := id.CIDRLabel.String()
	if !strings.HasPrefix(label, labels.LabelSourceCIDR) {
		log.WithFields(logrus.Fields{
			logfields.Identity: id.ID,
		}).Warning("BUG: Attempting to upsert non-CIDR identity")
		return
	}

	if prefix, err = netip.ParsePrefix(strings.TrimPrefix(label, labels.LabelSourceCIDR+":")); err != nil {
		log.WithFields(logrus.Fields{
			logfields.Identity: id.ID,
			logfields.Labels:   label,
		}).Warning("BUG: Attempting to upsert identity with bad CIDR label")
		return
	}
	return prefix, true
}

// upsertGeneratedIdentities unconditionally upserts 'newlyAllocatedIdentities'
// into the ipcache, then also upserts any CIDR identities in 'usedIdentities'
// that were not already upserted. If any 'usedIdentities' are upserted, these
// are counted separately as they may provide an indication of another logic
// error elsewhere in the codebase that is causing premature ipcache deletions.
//
// Deprecated: Prefer UpsertLabels() instead.
func (ipc *IPCache) upsertGeneratedIdentities(newlyAllocatedIdentities map[netip.Prefix]*identity.Identity, usedIdentities []*identity.Identity) {
	for prefix, id := range newlyAllocatedIdentities {
		ipc.Upsert(prefix.String(), nil, 0, nil, Identity{
			ID:     id.ID,
			Source: source.Generated,
		})
	}
	if len(usedIdentities) == 0 {
		return
	}

	toUpsert := make(map[netip.Prefix]*identity.Identity)
	ipc.mutex.Lock()
	for _, id := range usedIdentities {
		prefix, ok := cidrLabelToPrefix(id)
		if !ok {
			continue
		}
		prefixStr := prefix.String()
		entry, ok := ipc.ipToIdentityCache[prefixStr]
		if !ok {
			// We need this identity, but it was somehow deleted
			metrics.IPCacheErrorsTotal.WithLabelValues(
				metricTypeRecover, metricErrorUnexpected,
			).Inc()
			toUpsert[prefix] = id
			continue
		}

		if _, ok := newlyAllocatedIdentities[prefix]; ok {
			continue // already inserted above
		}

		// In case this entry is (co-)owned by the metadata API, we need to
		// store our legacy source here. Because our upsert source is "Generated",
		// we can assume that the existing entry source is equal or higher
		// precedence. Thus, we do not need to perform a full Upsert (which would
		// be rejected anyway), just update overwrittenLegacySource if needed.
		// This ensures that the entry is not removed once the higher-precedence
		// metadata resource owner is disassociated.
		switch {
		case entry.exclusivelyOwnedByLegacyAPI():
			// nothing to do
			continue
		case entry.exclusivelyOwnedByMetadataAPI():
			// mark entry as co-owned by both APIs
			entry.overwrittenLegacySource = source.Generated
			entry.modifiedByLegacyAPI = true
		case entry.ownedByLegacyAndMetadataAPI():
			// already co-owned, only update the legacy source if necessary
			if source.AllowOverwrite(entry.overwrittenLegacySource, source.Generated) {
				entry.overwrittenLegacySource = source.Generated
			}
		}

		ipc.ipToIdentityCache[prefixStr] = entry
	}
	ipc.mutex.Unlock()
	for prefix, id := range toUpsert {
		ipc.Upsert(prefix.String(), nil, 0, nil, Identity{
			ID:     id.ID,
			Source: source.Generated,
		})
	}
}

func (ipc *IPCache) releaseCIDRIdentities(ctx context.Context, prefixes []netip.Prefix) {
	// Create a critical section for identity release + removal from ipcache.
	// Otherwise, it's possible to trigger the following race condition:
	//
	// Goroutine 1                | Goroutine 2
	// releaseCIDRIdentities()    | AllocateCIDRs()
	// -> Release(..., id, ...)   |
	//                            | -> allocate(...)
	//                            | -> ipc.upsertGeneratedIdentities(...)
	// -> ipc.deleteLocked(...)   |
	//
	// In this case, the expectation from Goroutine 2 is that an identity
	// is allocated and that identity is in the ipcache, but the result
	// is that the identity is allocated but the ipcache entry is missing.
	ipc.Lock()
	defer ipc.Unlock()

	toDelete := make([]netip.Prefix, 0, len(prefixes))
	deletedIDs := make(map[identity.NumericIdentity]labels.LabelArray, len(prefixes))
	for _, prefix := range prefixes {
		lbls := labels.GetCIDRLabels(prefix)
		id := ipc.IdentityAllocator.LookupIdentity(ctx, lbls)
		if id == nil && option.Config.PolicyCIDRMatchesNodes() {
			// Hack for node-cidr feature.
			// We need to look up, exactly, the labels created during AllocateCIDRs(). Which we don't actually
			// know, since it might be a "normal" CIDR identity *or* a remote-node identity.
			//
			// So, if we don't find an identity for the CIDR label-set, and the node-cidr feature is enabled, then try
			// again with the set of labels for nodes.
			//
			// This can go away when CIDR identity restoration transitions to the UpsertLabels() api.
			lbls.MergeLabels(labels.LabelRemoteNode)
			lbls = lbls.Remove(labels.LabelWorld)
			lbls = lbls.Remove(labels.LabelWorldIPv4)
			lbls = lbls.Remove(labels.LabelWorldIPv6)
			id = ipc.IdentityAllocator.LookupIdentity(ctx, lbls)
		}
		if id == nil {
			log.Errorf("Unable to find identity of previously used CIDR %s", prefix.String())
			continue
		}

		released, err := ipc.IdentityAllocator.Release(ctx, id, false)
		if err != nil {
			log.WithFields(logrus.Fields{
				logfields.Identity: id,
				logfields.CIDR:     prefix,
			}).WithError(err).Warning("Unable to release CIDR identity. Ignoring error. Identity may be leaked")
		}
		if released {
			deletedIDs[id.ID] = id.LabelArray
			toDelete = append(toDelete, prefix)
		}
	}

	for _, prefix := range toDelete {
		ipc.deleteLocked(prefix.String(), source.Generated)
	}
	// Remove any deleted identities from the policy engine.
	ipc.UpdatePolicyMaps(ctx, nil, deletedIDs)
}

// ReleaseCIDRIdentitiesByCIDR releases the identities of a list of CIDRs.
// When the last use of the identity is released, the ipcache entry is deleted.
//
// Deprecated: Prefer RemoveLabels() or RemoveIdentity() instead.
func (ipc *IPCache) ReleaseCIDRIdentitiesByCIDR(prefixes []netip.Prefix) {
	ipc.deferredPrefixRelease.enqueue(prefixes, "cidr-prefix-release")
}
