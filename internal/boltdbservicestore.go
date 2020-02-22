package internal

import (
	"fmt"

	"github.com/boltdb/bolt"
	"github.com/vmihailenco/msgpack/v4"
)

var boltDBBucketName = []byte("services")

type BoltDBServiceStore struct {
	db *bolt.DB
}

func NewBoltDBServiceStore(path string) (*BoltDBServiceStore, error) {
	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		_, err = tx.CreateBucketIfNotExists(boltDBBucketName)
		return err
	})
	if err != nil {
		return nil, err
	}

	return &BoltDBServiceStore{
		db: db,
	}, nil
}

func (s *BoltDBServiceStore) addService(service *Service) error {
	serviceBytes, err := msgpack.Marshal(service)
	if err != nil {
		return err
	}

	err = s.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(boltDBBucketName).Put([]byte(service.ServiceURL), serviceBytes)
	})

	return err
}

func (s *BoltDBServiceStore) getServices() ([]*Service, error) {
	var serviceSlice []*Service

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(boltDBBucketName)
		c := b.Cursor()

		serviceSlice = make([]*Service, b.Stats().KeyN)

		var err error
		i := 0
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var service Service
			err = msgpack.Unmarshal(v, &service)
			if err != nil {
				return err
			}
			serviceSlice[i] = &service
			i++
		}

		return nil
	})
	if err != nil {
		return serviceSlice, err
	}

	return serviceSlice, nil
}

func (s *BoltDBServiceStore) Close() error {
	return s.db.Close()
}

func (s *BoltDBServiceStore) String() string {
	return fmt.Sprintf("boltb (%s)", s.db.Path())
}
