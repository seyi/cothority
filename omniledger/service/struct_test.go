package service

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	bolt "github.com/coreos/bbolt"
	"github.com/stretchr/testify/require"
)

var testName = []byte("coll1")

func TestCollectionDBStrange(t *testing.T) {
	tmpDB, err := ioutil.TempFile("", "tmpDB")
	require.Nil(t, err)
	tmpDB.Close()
	defer os.Remove(tmpDB.Name())

	db, err := bolt.Open(tmpDB.Name(), 0600, nil)
	require.Nil(t, err)

	cdb := newCollectionDB(db, testName)
	key := []byte("first")
	value := []byte("value")
	contract := "mycontract"
	err = cdb.Store(&StateChange{
		StateAction: Create,
		InstanceID:  key,
		Value:       value,
		ContractID:  []byte(contract),
	})
	require.Nil(t, err)
	v, c, _, err := cdb.GetValues([]byte("first"))
	require.Nil(t, err)
	require.Equal(t, value, v)
	require.Equal(t, contract, c)
}

func TestCollectionDB(t *testing.T) {
	kvPairs := 16

	tmpDB, err := ioutil.TempFile("", "tmpDB")
	require.Nil(t, err)
	tmpDB.Close()
	defer os.Remove(tmpDB.Name())

	db, err := bolt.Open(tmpDB.Name(), 0600, nil)
	require.Nil(t, err)

	cdb := newCollectionDB(db, testName)
	pairs := map[string]string{}
	myContract := "myContract"
	for i := 0; i < kvPairs; i++ {
		pairs[fmt.Sprintf("Key%d", i)] = fmt.Sprintf("value%d", i)
	}

	// Store all key/value pairs
	for k, v := range pairs {
		sc := &StateChange{
			StateAction: Create,
			InstanceID:  []byte(k),
			Value:       []byte(v),
			ContractID:  []byte(myContract),
		}
		require.Nil(t, cdb.Store(sc))
	}

	// Verify it's all there
	for c, v := range pairs {
		stored, contract, _, err := cdb.GetValues([]byte(c))
		require.Nil(t, err)
		require.Equal(t, v, string(stored))
		require.Equal(t, myContract, contract)
	}

	// Get a new db handler
	cdb2 := newCollectionDB(db, testName)

	// Verify it's all there
	for c, v := range pairs {
		stored, _, _, err := cdb2.GetValues([]byte(c))
		require.Nil(t, err)
		require.Equal(t, v, string(stored))
	}

	// Update
	for k, v := range pairs {
		pairs[k] = v + "-2"
	}
	for k, v := range pairs {
		sc := &StateChange{
			StateAction: Update,
			InstanceID:  []byte(k),
			Value:       []byte(v),
			ContractID:  []byte(myContract),
		}
		require.Nil(t, cdb2.Store(sc), k)
	}
	for k, v := range pairs {
		stored, contract, _, err := cdb2.GetValues([]byte(k))
		require.Nil(t, err)
		require.Equal(t, v, string(stored))
		require.Equal(t, myContract, contract)
	}

	// Delete
	for c := range pairs {
		sc := &StateChange{
			StateAction: Remove,
			InstanceID:  []byte(c),
			ContractID:  []byte(myContract),
		}
		require.Nil(t, cdb2.Store(sc))
	}
	for c := range pairs {
		_, _, _, err := cdb2.GetValues([]byte(c))
		require.NotNil(t, err, c)
	}
}

// TODO: Test good case, bad add case, bad remove case
func TestCollectionDBtryHash(t *testing.T) {
	tmpDB, err := ioutil.TempFile("", "tmpDB")
	require.Nil(t, err)
	tmpDB.Close()
	defer os.Remove(tmpDB.Name())

	db, err := bolt.Open(tmpDB.Name(), 0600, nil)
	require.Nil(t, err)

	cdb := newCollectionDB(db, testName)
	scs := []StateChange{{
		StateAction: Create,
		InstanceID:  []byte("key1"),
		ContractID:  []byte("kind1"),
		Value:       []byte("value1"),
	},
		{
			StateAction: Create,
			InstanceID:  []byte("key2"),
			ContractID:  []byte("kind2"),
			Value:       []byte("value2"),
		},
	}
	mrTrial, err := cdb.tryHash(scs)
	require.Nil(t, err)
	_, _, _, err = cdb.GetValues([]byte("key1"))
	require.EqualError(t, err, "no match found")
	_, _, _, err = cdb.GetValues([]byte("key2"))
	require.EqualError(t, err, "no match found")
	cdb.Store(&scs[0])
	cdb.Store(&scs[1])
	mrReal := cdb.RootHash()
	require.Equal(t, mrTrial, mrReal)
}
