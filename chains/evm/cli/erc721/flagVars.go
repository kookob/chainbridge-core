package erc721

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/kookob/chainbridge-core/crypto/secp256k1"
	"github.com/kookob/chainbridge-core/types"
)

// flag vars
var (
	Erc721Address  string
	DstAddress     string
	TokenId        string
	Metadata       string
	Recipient      string
	Bridge         string
	DestionationID string
	ResourceID     string
	Minter         string
)

// processed flag vars
var (
	erc721Addr    common.Address
	dstAddress    common.Address
	tokenId       *big.Int
	recipientAddr common.Address
	bridgeAddr    common.Address
	destinationID int
	resourceId    types.ResourceID
	minterAddr    common.Address
)

// global flags
var (
	url           string
	gasLimit      uint64
	gasPrice      *big.Int
	senderKeyPair *secp256k1.Keypair
	err           error
)
