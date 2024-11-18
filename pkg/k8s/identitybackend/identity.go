// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Cilium

package identitybackend

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"

	"github.com/cilium/cilium/pkg/allocator"
	cacheKey "github.com/cilium/cilium/pkg/identity/key"
	"github.com/cilium/cilium/pkg/idpool"
	k8sConst "github.com/cilium/cilium/pkg/k8s/apis/cilium.io"
	v2 "github.com/cilium/cilium/pkg/k8s/apis/cilium.io/v2"
	clientset "github.com/cilium/cilium/pkg/k8s/client/clientset/versioned"
	"github.com/cilium/cilium/pkg/k8s/informer"
	k8sUtils "github.com/cilium/cilium/pkg/k8s/utils"
	"github.com/cilium/cilium/pkg/kvstore"
	"github.com/cilium/cilium/pkg/labels"
	"github.com/cilium/cilium/pkg/logging"
	"github.com/cilium/cilium/pkg/logging/logfields"
	"github.com/cilium/cilium/pkg/rate"
)

var (
	log = logging.DefaultLogger.WithField(logfields.LogSubsys, "crd-allocator")
)

const (
	// HeartBeatAnnotation is an annotation applied by the operator to indicate
	// that a CiliumIdentity has been marked for deletion.
	HeartBeatAnnotation = "io.cilium.heartbeat"

	k8sPrefix               = labels.LabelSourceK8s + ":"
	k8sNamespaceLabelPrefix = labels.LabelSourceK8s + ":" + k8sConst.PodNamespaceMetaLabels + labels.PathDelimiter

	// byKeyIndex is the name of the index of the identities by key.
	byKeyIndex = "by-key-index"
)

func NewCRDBackend(c CRDBackendConfiguration) (allocator.Backend, error) {
	return &crdBackend{CRDBackendConfiguration: c}, nil
}

type CRDBackendConfiguration struct {
	Store   cache.Indexer
	Client  clientset.Interface
	KeyFunc func(map[string]string) allocator.AllocatorKey
}

type crdBackend struct {
	CRDBackendConfiguration
}

func (c *crdBackend) DeleteAllKeys(ctx context.Context) {
}

func (c *crdBackend) DeleteID(ctx context.Context, id idpool.ID) error {
	return c.Client.CiliumV2().CiliumIdentities().Delete(ctx, id.String(), metav1.DeleteOptions{})
}

// SanitizeK8sLabels strips the 'k8s:' prefix in the labels generated by
// AllocatorKey.GetAsMap (when the key is k8s labels). In the CRD identity case
// we map the labels directly to the ciliumidentity CRD instance, and
// kubernetes does not allow ':' in the name of the label. These labels are not
// the canonical labels of the identity, but used to ease interaction with the
// CRD object.
func SanitizeK8sLabels(old map[string]string) (selected, skipped map[string]string) {
	skipped = make(map[string]string, len(old))
	selected = make(map[string]string, len(old))
	for k, v := range old {
		// Skip non-k8s labels.
		// Skip synthesized labels for k8s namespace labels, since they contain user input which can result in the label
		// name being longer than 63 characters.
		if !strings.HasPrefix(k, k8sPrefix) || strings.HasPrefix(k, k8sNamespaceLabelPrefix) {
			skipped[k] = v
			continue // skip non-k8s labels
		}
		k = strings.TrimPrefix(k, k8sPrefix) // k8s: is redundant
		selected[k] = v
	}
	return selected, skipped
}

// AllocateID will create an identity CRD, thus creating the identity for this
// key-> ID mapping.
// Note: the lock field is not supported with the k8s CRD allocator.
// Returns an allocator key with the cilium identity stored in it.
func (c *crdBackend) AllocateID(ctx context.Context, id idpool.ID, key allocator.AllocatorKey) (allocator.AllocatorKey, error) {
	selectedLabels, skippedLabels := SanitizeK8sLabels(key.GetAsMap())
	log.WithField(logfields.Labels, skippedLabels).Info("Skipped non-kubernetes labels when labelling ciliumidentity. All labels will still be used in identity determination")

	identity := &v2.CiliumIdentity{
		ObjectMeta: metav1.ObjectMeta{
			Name:   id.String(),
			Labels: selectedLabels,
		},
		SecurityLabels: key.GetAsMap(),
	}

	ci, err := c.Client.CiliumV2().CiliumIdentities().Create(ctx, identity, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}
	return key.PutValue(cacheKey.MetadataKeyBackendKey, ci), nil
}

func (c *crdBackend) AllocateIDIfLocked(ctx context.Context, id idpool.ID, key allocator.AllocatorKey, lock kvstore.KVLocker) (allocator.AllocatorKey, error) {
	return c.AllocateID(ctx, id, key)
}

// AcquireReference acquires a reference to the identity.
func (c *crdBackend) AcquireReference(ctx context.Context, id idpool.ID, key allocator.AllocatorKey, lock kvstore.KVLocker) error {
	// For CiliumIdentity-based allocation, the reference counting is
	// handled via CiliumEndpoint. Any CiliumEndpoint referring to a
	// CiliumIdentity will keep the CiliumIdentity alive. However,
	// there is a brief window where a CiliumEndpoint may not exist
	// for a given CiliumIdentity (according to the operator), in
	// which case the operator marks the CiliumIdentity for deletion.
	// This checks to see if the CiliumIdentity has been marked for
	// deletion and removes the mark so that the CiliumIdentity can
	// be safely used.
	//
	// NOTE: A race against using a CiliumIdentity that might otherwise
	// be (immediately) deleted is prevented by the operator logic that
	// validates the ResourceVersion of the CiliumIdentity before deleting
	// it. If a CiliumIdentity does (eventually) get deleted by the
	// operator, the agent will then have a chance to recreate it.
	var (
		ts string
		ok bool
	)
	// check to see if the cached copy of the identity
	// has the annotation
	ci, exists, err := c.getById(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		// fall back to the key stored in the allocator key. If it's not present
		// then return the error.
		ci, ok = key.Value(cacheKey.MetadataKeyBackendKey).(*v2.CiliumIdentity)
		if !ok {
			return fmt.Errorf("identity (id:%q,key:%q) does not exist", id, key)
		}
	}

	ts, ok = ci.Annotations[HeartBeatAnnotation]
	if ok {
		log.WithField(logfields.Identity, ci).Infof("Identity marked for deletion (at %s); attempting to unmark it", ts)
		ci = ci.DeepCopy()
		delete(ci.Annotations, HeartBeatAnnotation)
		_, err = c.Client.CiliumV2().CiliumIdentities().Update(ctx, ci, metav1.UpdateOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *crdBackend) RunLocksGC(_ context.Context, _ map[string]kvstore.Value) (map[string]kvstore.Value, error) {
	return nil, nil
}

func (c *crdBackend) RunGC(context.Context, *rate.Limiter, map[string]uint64, idpool.ID, idpool.ID) (map[string]uint64, *allocator.GCStats, error) {
	return nil, nil, nil
}

// UpdateKey refreshes the reference that this node is using this key->ID
// mapping. It assumes that the identity already exists but will recreate it if
// reliablyMissing is true.
// Note: the lock field is not supported with the k8s CRD allocator.
func (c *crdBackend) UpdateKey(ctx context.Context, id idpool.ID, key allocator.AllocatorKey, reliablyMissing bool) error {
	err := c.AcquireReference(ctx, id, key, nil)
	if err == nil {
		log.WithFields(logrus.Fields{
			logfields.Identity: id,
			logfields.Labels:   key,
		}).Debug("Acquired reference for identity")
		return nil
	}

	// The CRD (aka the master key) is missing. Try to recover by recreating it
	// if reliablyMissing is set.
	log.WithError(err).WithFields(logrus.Fields{
		logfields.Identity: id,
		logfields.Labels:   key,
	}).Warning("Unable update CRD identity information with a reference for this node")

	if reliablyMissing {
		// Recreate a missing master key
		if _, err = c.AllocateID(ctx, id, key); err != nil {
			return fmt.Errorf("Unable recreate missing CRD identity %q->%q: %w", key, id, err)
		}

		return nil
	}

	return err
}

func (c *crdBackend) UpdateKeyIfLocked(ctx context.Context, id idpool.ID, key allocator.AllocatorKey, reliablyMissing bool, lock kvstore.KVLocker) error {
	return c.UpdateKey(ctx, id, key, reliablyMissing)
}

// Lock does not return a lock object. Locking is not supported with the k8s
// CRD allocator. It is here to meet interface requirements.
func (c *crdBackend) Lock(ctx context.Context, key allocator.AllocatorKey) (kvstore.KVLocker, error) {
	return &crdLock{}, nil
}

type crdLock struct{}

// Unlock does not unlock a lock object. Locking is not supported with the k8s
// CRD allocator. It is here to meet interface requirements.
func (c *crdLock) Unlock(ctx context.Context) error {
	return nil
}

// Comparator does nothing. Locking is not supported with the k8s
// CRD allocator. It is here to meet interface requirements.
func (c *crdLock) Comparator() interface{} {
	return nil
}

// get returns the identity found for the given set of labels.
// In the case of duplicate entries, return an identity entry
// from a sorted list.
func (c *crdBackend) get(ctx context.Context, key allocator.AllocatorKey) *v2.CiliumIdentity {
	if c.Store == nil {
		return nil
	}

	identities, err := c.Store.ByIndex(byKeyIndex, key.GetKey())
	if err != nil || len(identities) == 0 {
		return nil
	}

	sort.Slice(identities, func(i, j int) bool {
		left, ok := identities[i].(*v2.CiliumIdentity)
		if !ok {
			return false
		}

		right, ok := identities[j].(*v2.CiliumIdentity)
		if !ok {
			return false
		}

		return left.CreationTimestamp.Before(&right.CreationTimestamp)
	})

	for _, identityObject := range identities {
		identity, ok := identityObject.(*v2.CiliumIdentity)
		if !ok {
			return nil
		}

		if reflect.DeepEqual(identity.SecurityLabels, key.GetAsMap()) {
			return identity
		}
	}
	return nil
}

// Get returns the first ID which is allocated to a key in the identity CRDs in
// kubernetes.
// Note: the lock field is not supported with the k8s CRD allocator.
func (c *crdBackend) Get(ctx context.Context, key allocator.AllocatorKey) (idpool.ID, error) {
	identity := c.get(ctx, key)
	if identity == nil {
		return idpool.NoID, nil
	}

	id, err := strconv.ParseUint(identity.Name, 10, 64)
	if err != nil {
		return idpool.NoID, fmt.Errorf("unable to parse value '%s': %w", identity.Name, err)
	}

	return idpool.ID(id), nil
}

func (c *crdBackend) GetIfLocked(ctx context.Context, key allocator.AllocatorKey, lock kvstore.KVLocker) (idpool.ID, error) {
	return c.Get(ctx, key)
}

// getById fetches the identities from the local store. Returns a nil `err` and
// false `exists` if an Identity is not found for the given `id`.
func (c *crdBackend) getById(ctx context.Context, id idpool.ID) (idty *v2.CiliumIdentity, exists bool, err error) {
	if c.Store == nil {
		return nil, false, fmt.Errorf("store is not available yet")
	}

	identityTemplate := &v2.CiliumIdentity{
		ObjectMeta: metav1.ObjectMeta{
			Name: id.String(),
		},
	}

	obj, exists, err := c.Store.Get(identityTemplate)
	if err != nil {
		return nil, exists, err
	}
	if !exists {
		return nil, exists, nil
	}

	identity, ok := obj.(*v2.CiliumIdentity)
	if !ok {
		return nil, false, fmt.Errorf("invalid object %T", obj)
	}
	return identity, true, nil
}

// GetByID returns the key associated with an ID. Returns nil if no key is
// associated with the ID.
// Note: the lock field is not supported with the k8s CRD allocator.
func (c *crdBackend) GetByID(ctx context.Context, id idpool.ID) (allocator.AllocatorKey, error) {
	identity, exists, err := c.getById(ctx, id)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, nil
	}

	return c.KeyFunc(identity.SecurityLabels), nil
}

// Release dissociates this node from using the identity bound to the given ID.
// When an identity has no references it may be garbage collected.
func (c *crdBackend) Release(ctx context.Context, id idpool.ID, key allocator.AllocatorKey) (err error) {
	// For CiliumIdentity-based allocation, the reference counting is
	// handled via CiliumEndpoint. Any CiliumEndpoint referring to a
	// CiliumIdentity will keep the CiliumIdentity alive. No action is
	// needed to release the reference here.
	return nil
}

func getIdentitiesByKeyFunc(keyFunc func(map[string]string) allocator.AllocatorKey) func(obj interface{}) ([]string, error) {
	return func(obj interface{}) ([]string, error) {
		if identity, ok := obj.(*v2.CiliumIdentity); ok {
			return []string{keyFunc(identity.SecurityLabels).GetKey()}, nil
		}
		return []string{}, fmt.Errorf("object other than CiliumIdentity was pushed to the store")
	}
}

func (c *crdBackend) ListIDs(ctx context.Context) (identityIDs []idpool.ID, err error) {
	if c.Store == nil {
		return nil, fmt.Errorf("store is not available yet")
	}

	for _, identity := range c.Store.List() {
		idParsed, err := strconv.ParseUint(identity.(*v2.CiliumIdentity).Name, 10, 64)
		if err != nil {
			log.WithField(logfields.Identity, identity.(*v2.CiliumIdentity).Name).Warn("Cannot parse identity ID")
			continue
		}
		identityIDs = append(identityIDs, idpool.ID(idParsed))
	}
	return identityIDs, err
}

func (c *crdBackend) ListAndWatch(ctx context.Context, handler allocator.CacheMutations, stopChan chan struct{}) {
	c.Store = cache.NewIndexer(
		cache.DeletionHandlingMetaNamespaceKeyFunc,
		cache.Indexers{byKeyIndex: getIdentitiesByKeyFunc(c.KeyFunc)})
	identityInformer := informer.NewInformerWithStore(
		k8sUtils.ListerWatcherFromTyped[*v2.CiliumIdentityList](c.Client.CiliumV2().CiliumIdentities()),
		&v2.CiliumIdentity{},
		0,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				if identity, ok := obj.(*v2.CiliumIdentity); ok {
					if id, err := strconv.ParseUint(identity.Name, 10, 64); err == nil {
						handler.OnUpsert(idpool.ID(id), c.KeyFunc(identity.SecurityLabels))
					}
				}
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				if oldIdentity, ok := oldObj.(*v2.CiliumIdentity); ok {
					if newIdentity, ok := newObj.(*v2.CiliumIdentity); ok {
						if oldIdentity.DeepEqual(newIdentity) {
							return
						}
						if id, err := strconv.ParseUint(newIdentity.Name, 10, 64); err == nil {
							handler.OnUpsert(idpool.ID(id), c.KeyFunc(newIdentity.SecurityLabels))
						}
					}
				}
			},
			DeleteFunc: func(obj interface{}) {
				// The delete event is sometimes for items with unknown state that are
				// deleted anyway.
				if deleteObj, isDeleteObj := obj.(cache.DeletedFinalStateUnknown); isDeleteObj {
					obj = deleteObj.Obj
				}

				if identity, ok := obj.(*v2.CiliumIdentity); ok {
					if id, err := strconv.ParseUint(identity.Name, 10, 64); err == nil {
						handler.OnDelete(idpool.ID(id), c.KeyFunc(identity.SecurityLabels))
					}
				} else {
					log.Debugf("Ignoring unknown delete event %#v", obj)
				}
			},
		},
		nil,
		c.Store,
	)

	go func() {
		if ok := cache.WaitForCacheSync(stopChan, identityInformer.HasSynced); ok {
			handler.OnListDone()
		}
	}()

	identityInformer.Run(stopChan)
}
