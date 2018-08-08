package service

import (
	"time"

	"github.com/dedis/cothority/omniledger/collection"
	"github.com/dedis/cothority/omniledger/darc"
	"github.com/dedis/cothority/skipchain"
	"github.com/dedis/onet"
)

// PROTOSTART
// type :skipchain.SkipBlockID:bytes
// type :darc.ID:bytes
// type :Arguments:[]Argument
// type :Instructions:[]Instruction
// type :ClientTransactions:[]ClientTransaction
// package omniledger;
// import "skipchain.proto";
// import "onet.proto";
// import "darc.proto";
// import "collection.proto";
//
// option java_package = "ch.epfl.dedis.proto";
// option java_outer_classname = "OmniLedgerProto";

// DataHeader is the data passed to the Skipchain
type DataHeader struct {
	// CollectionRoot is the root of the merkle tree of the colleciton after
	// applying the valid transactions.
	CollectionRoot []byte
	// ClientTransactionHash is the sha256 hash of all the transactions in the body
	ClientTransactionHash []byte
	// StateChangesHash is the sha256 of all the stateChanges occuring through the
	// clientTransactions.
	StateChangesHash []byte
	// Timestamp is a unix timestamp in nanoseconds.
	Timestamp int64
}

// DataBody is stored in the body of the skipblock but is not hashed. This reduces
// the proof needed for a key/value pair.
type DataBody struct {
	Transactions ClientTransactions
}

// ***
// These are the messages used in the API-calls
// ***

// CreateGenesisBlock asks the cisc-service to set up a new skipchain.
type CreateGenesisBlock struct {
	// Version of the protocol
	Version Version
	// Roster defines which nodes participate in the skipchain.
	Roster onet.Roster
	// GenesisDarc defines who is allowed to write to this skipchain.
	GenesisDarc darc.Darc
	// BlockInterval in int64.
	BlockInterval time.Duration
}

// CreateGenesisBlockResponse holds the genesis-block of the new skipchain.
type CreateGenesisBlockResponse struct {
	// Version of the protocol
	Version Version
	// Skipblock of the created skipchain or empty if there was an error.
	Skipblock *skipchain.SkipBlock
}

// AddTxRequest requests to apply a new transaction to the ledger.
type AddTxRequest struct {
	// Version of the protocol
	Version Version
	// SkipchainID is the hash of the first skipblock
	SkipchainID skipchain.SkipBlockID
	// Transaction to be applied to the kv-store
	Transaction ClientTransaction
	// How many block-intervals to wait for inclusion -
	// missing value or 0 means return immediately.
	InclusionWait int `protobuf:"opt"`
}

// AddTxResponse is the reply after an AddTxRequest is finished.
type AddTxResponse struct {
	// Version of the protocol
	Version Version
}

// GetProof returns the proof that the given key is in the collection.
type GetProof struct {
	// Version of the protocol
	Version Version
	// Key is the key we want to look up
	Key []byte
	// ID is any block that is known to us in the skipchain, can be the genesis
	// block or any later block. The proof returned will be starting at this block.
	ID skipchain.SkipBlockID
}

// GetProofResponse can be used together with the Genesis block to proof that
// the returned key/value pair is in the collection.
type GetProofResponse struct {
	// Version of the protocol
	Version Version
	// Proof contains everything necessary to prove the inclusion
	// of the included key/value pair given a genesis skipblock.
	Proof Proof
}

// ChainConfig stores all the configuration information for one skipchain. It will
// be stored under the key "GenesisDarcID || OneNonce", in the collections. The
// GenesisDarcID is the value of GenesisReferenceID.
type ChainConfig struct {
	BlockInterval time.Duration
	Roster        onet.Roster
}

// Proof represents everything necessary to verify a given
// key/value pair is stored in a skipchain. The proof is in three parts:
//   1. InclusionProof proofs the presence or absence of the key. In case of
//   the key being present, the value is included in the proof
//   2. Latest is used to verify the merkle tree root used in the collection-proof
//   is stored in the latest skipblock
//   3. Links proves that the latest skipblock is part of the skipchain
//
// This Structure could later be moved to cothority/skipchain.
type Proof struct {
	// InclusionProof is the deserialized InclusionProof
	InclusionProof collection.Proof
	// Providing the latest skipblock to retrieve the Merkle tree root.
	Latest skipchain.SkipBlock
	// Proving the path to the latest skipblock. The first ForwardLink has an
	// empty-sliced `From` and the genesis-block in `To`, together with the
	// roster of the genesis-block in the `NewRoster`.
	Links []skipchain.ForwardLink
}

// Instruction holds only one of Spawn, Invoke, or Delete
type Instruction struct {
	// InstanceID is either the object that can spawn a new object, or a the object
	// that will be invoked or deleted.
	InstanceID InstanceID
	// Nonce is monotonically increasing with regard to the Darc controlling
	// access to the instance. It is used to prevent replay attacks.
	// The client has to track what the next nonce should be for a given instance.
	Nonce Nonce
	// Index and length prevent a leader from censoring specific instructions from
	// a client and still keep the other instructions valid.
	// Index is relative to the beginning of the clientTransaction.
	Index int
	// Length is the total number of instructions in this clientTransaction
	Length int
	// Spawn creates a new object.
	Spawn *Spawn
	// Invoke calls a method of an existing object.
	Invoke *Invoke
	// Delete removes the given object.
	Delete *Delete
	// Signatures that will be used while attempting to verify the Darc read request.
	Signatures []darc.Signature
}

// Spawn is called upon an existing object that will spawn a new object.
type Spawn struct {
	// ContractID represents the kind of contract that needs to be spawn.
	ContractID string
	// Args holds all data necessary to spawn the new object.
	Args Arguments
}

// Invoke calls a method of an existing object which will update its internal
// state.
type Invoke struct {
	// Command is object specific and interpreted by the object.
	Command string
	// Args holds all data necessary for the successful execution of the command.
	Args Arguments
}

// Delete removes the object.
type Delete struct {
}

// Argument is a name/value pair that will be passed to the object.
type Argument struct {
	// Name can be any name recognized by the object.
	Name string
	// Value must be binary marshalled
	Value []byte
}

// ClientTransaction is a slice of Instructions that will be applied in order.
// If any of the instructions fails, none of them will be applied.
type ClientTransaction struct {
	Instructions Instructions
}

// StateChange is one new state that will be applied to the collection.
type StateChange struct {
	// StateAction can be any of Create, Update, Remove
	StateAction StateAction
	// InstanceID of the state to change
	InstanceID []byte
	// ContractID points to the contract that can interpret the value
	ContractID []byte
	// Value is the data needed by the contract
	Value []byte
	// DarcID is the Darc controlling access to this key.
	DarcID darc.ID
}

// Coin is a generic structure holding any type of coin. Coins are defined
// by a genesis coin object that is unique for each type of coin.
type Coin struct {
	// Name points to the genesis object of that coin.
	Name InstanceID
	// Value is the total number of coins of that type.
	Value uint64
}
