package db

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/dgraph-io/badger/v3"
)

const dbPath = ".db"

type DB struct {
	db *badger.DB
}

func InitDB() (*DB, error) {
	opts := badger.DefaultOptions(dbPath)
	opts.Logger = nil

	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}

	return &DB{
		db: db,
	}, nil
}

func (d *DB) Close() error {
	return d.db.Close()
}

func (d *DB) Set(key []byte, value interface{}) error {
	var valueBytes bytes.Buffer
	err := json.NewEncoder(&valueBytes).Encode(value)
	if err != nil {
		return err
	}

	return d.db.Update(
		func(txn *badger.Txn) error {
			err := txn.Set(key, valueBytes.Bytes())
			if err != nil {
				return err
			}

			return nil
		},
	)
}

func (d *DB) Get(key []byte, v interface{}) error {
	if v == nil {
		return fmt.Errorf("v is not set")
	}

	var data []byte

	err := d.db.View(
		func(txn *badger.Txn) error {
			item, err := txn.Get(key)
			if err != nil {
				return err
			}

			err = item.Value(
				func(val []byte) error {
					data = append(data, val...)

					return nil
				},
			)
			if err != nil {
				return err
			}

			return nil
		},
	)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, v)
}

func (d *DB) GetFirstWherePrefix(prefix []byte, v interface{}) error {
	if v == nil {
		return fmt.Errorf("v is not set")
	}

	var data []byte

	err := d.db.View(
		func(txn *badger.Txn) error {
			it := txn.NewIterator(
				badger.IteratorOptions{
					PrefetchValues: false,
					Reverse:        false,
					AllVersions:    false,
				},
			)
			defer it.Close()

			for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
				item := it.Item()

				err := item.Value(
					func(val []byte) error {
						if data == nil {
							data = append(data, val...)
						}
						return nil
					},
				)
				if err != nil {
					return err
				}

				it.Close()
			}
			return nil
		},
	)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, v)
}

func (d *DB) Drop(prefix []byte) error {
	return d.db.DropPrefix(prefix)
}
