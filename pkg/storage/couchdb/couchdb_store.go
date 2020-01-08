/*
Copyright SecureKey Technologies Inc. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package couchdb

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/flimzy/kivik"
	_ "github.com/go-kivik/couchdb" // The CouchDB driver

	"github.com/trustbloc/edge-store/pkg/storage"
)

// Provider represents an CouchDB implementation of the storage.Provider interface
type Provider struct {
	hostURL       string
	couchDBClient *kivik.Client
}

// NewProvider instantiates Provider
func NewProvider(hostURL string) (*Provider, error) {
	if hostURL == "" {
		return nil, errors.New("hostURL for new CouchDB provider can't be blank")
	}

	client, err := kivik.New(context.Background(), "couch", hostURL)
	if err != nil {
		return nil, err
	}

	return &Provider{hostURL: hostURL, couchDBClient: client}, nil
}

// OpenStore opens and returns a store for the given name.
func (p Provider) OpenStore(name string) (storage.Store, error) {
	exists, err := p.couchDBClient.DBExists(context.Background(), name)
	if err != nil {
		return nil, err
	}

	if exists {
		db, getDBErr := p.couchDBClient.DB(context.Background(), name)
		if getDBErr != nil {
			return nil, getDBErr
		}

		return couchDBStore{db: db}, nil
	}

	err = p.couchDBClient.CreateDB(context.Background(), name)
	if err != nil {
		return nil, err
	}

	db, err := p.couchDBClient.DB(context.Background(), name)
	if err != nil {
		return nil, err
	}

	return couchDBStore{db: db}, nil
}

// CloseStore closes a previously opened store.
// CouchDB and Kivik don't really have "close" mechanisms for individual databases, so there's nothing to do here.
func (p Provider) CloseStore(name string) error {
	return nil
}

// Close closes the provider.
func (p Provider) Close() error {
	p.couchDBClient = nil
	return nil
}

// StoreExists tells you whether a store with the given name is already open.
func (p Provider) StoreExists(name string) (bool, error) {
	exists, err := p.couchDBClient.DBExists(context.Background(), name)
	return exists, err
}

type couchDBStore struct {
	db *kivik.DB
}

func (c couchDBStore) Put(k string, v []byte) error {
	_, err := c.db.Put(context.Background(), k, v)
	if err != nil {
		return err
	}

	return nil
}

func (c couchDBStore) Get(k string) ([]byte, error) {
	row, err := c.db.Get(context.Background(), k)
	if err != nil {
		if err.Error() == "Not Found: missing" {
			return nil, storage.ErrValueNotFound
		}

		return nil, err
	}

	destinationData := make(map[string]interface{})

	err = row.ScanDoc(&destinationData)
	if err != nil {
		return nil, err
	}

	// Stripping out the CouchDB fields
	delete(destinationData, "_id")
	delete(destinationData, "_rev")

	strippedJSON, err := json.Marshal(destinationData)
	if err != nil {
		return nil, err
	}

	return strippedJSON, nil
}
