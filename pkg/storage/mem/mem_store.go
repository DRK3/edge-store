/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mem

import (
	"github.com/trustbloc/edge-store/pkg/storage"
)

// Provider represents an memStore implementation of the storage.Provider interface.
// memStore is a simple DB that's stored in memory. Useful for demos or testing. Not designed to be performant.
type Provider struct {
	dbs map[string]*memStore
}

// NewProvider instantiates Provider
func NewProvider() *Provider {
	return &Provider{dbs: make(map[string]*memStore)}
}

// OpenStore opens and returns a store for the given name.
func (p *Provider) OpenStore(name string) (storage.Store, error) {
	store, exists := p.dbs[name]
	if !exists {
		return p.newMemStore(name), nil
	}

	return store, nil
}

func (p *Provider) newMemStore(name string) *memStore {
	store := memStore{db: make(map[string][]byte)}

	p.dbs[name] = &store

	return &store
}

// CloseStore closes a previously opened store.
func (p *Provider) CloseStore(name string) error {
	store, exists := p.dbs[name]
	if !exists {
		return storage.ErrStoreNotFound
	}

	delete(p.dbs, name)

	store.close()

	return nil
}

// Close closes the provider.
func (p *Provider) Close() error {
	for _, memStore := range p.dbs {
		memStore.db = make(map[string][]byte)
	}

	p.dbs = make(map[string]*memStore)

	return nil
}

// StoreExists tells you whether a store with the given name is already open.
func (p *Provider) StoreExists(name string) (bool, error) {
	_, exists := p.dbs[name]
	return exists, nil
}

type memStore struct {
	db map[string][]byte
}

// Put stores the given key-value pair in the store.
func (store *memStore) Put(k string, v []byte) error {
	store.db[k] = v

	return nil
}

// Get retrieves the value in the store associated with the given key.
func (store *memStore) Get(k string) ([]byte, error) {
	v, exists := store.db[k]
	if !exists {
		return nil, storage.ErrValueNotFound
	}

	return v, nil
}

func (store *memStore) close() {
	store.db = make(map[string][]byte)
}
