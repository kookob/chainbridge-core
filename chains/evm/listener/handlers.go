package listener

import (
	erc20Handler "github.com/ChainSafe/chainbridgev2/bindings/eth/bindings/ERC20Handler"
	erc721Handler "github.com/ChainSafe/chainbridgev2/bindings/eth/bindings/ERC721Handler"
	genericHandler "github.com/ChainSafe/chainbridgev2/bindings/eth/bindings/GenericHandler"
	"github.com/ChainSafe/chainbridgev2/chains/evm"
	"github.com/ChainSafe/chainbridgev2/relayer"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

func HandleErc20DepositedEvent(sourceID, destId uint8, nonce uint64, handlerContractAddress string, backend ChainClient) (relayer.XCMessager, error) {
	contract, err := erc20Handler.NewERC20HandlerCaller(common.HexToAddress(handlerContractAddress), backend)
	if err != nil {
		return nil, err
	}
	record, err := contract.GetDepositRecord(&bind.CallOpts{}, uint64(nonce), uint8(destId))
	if err != nil {
		return nil, err
	}

	return &evm.DefaultEVMMessage{
		Source:       sourceID,
		Destination:  destId,
		Type:         evm.FungibleTransfer,
		DepositNonce: nonce,
		ResourceId:   record.ResourceID,
		Payload: []interface{}{
			record.Amount.Bytes(),
			record.DestinationRecipientAddress,
		},
	}, nil
}

func HandleErc721DepositedEvent(sourceID, destId uint8, nonce uint64, handlerContractAddress string, backend ChainClient) (relayer.XCMessager, error) {
	contract, err := erc721Handler.NewERC721HandlerCaller(common.HexToAddress(handlerContractAddress), backend)
	if err != nil {
		return nil, err
	}
	record, err := contract.GetDepositRecord(&bind.CallOpts{}, uint64(nonce), uint8(destId))
	if err != nil {
		return nil, err
	}
	return &evm.DefaultEVMMessage{
		Source:       sourceID,
		Destination:  destId,
		Type:         evm.NonFungibleTransfer,
		DepositNonce: nonce,
		ResourceId:   record.ResourceID,
		Payload: []interface{}{
			record.TokenID.Bytes(),
			record.DestinationRecipientAddress,
			record.MetaData,
		},
	}, nil
}

func HandleGenericDepositedEvent(sourceID, destId uint8, nonce uint64, handlerContractAddress string, backend ChainClient) (relayer.XCMessager, error) {
	contract, err := genericHandler.NewGenericHandlerCaller(common.HexToAddress(handlerContractAddress), backend)
	if err != nil {
		return nil, err
	}
	record, err := contract.GetDepositRecord(&bind.CallOpts{}, uint64(nonce), uint8(destId))
	if err != nil {
		return nil, err
	}
	return &evm.DefaultEVMMessage{
		Source:       sourceID,
		Destination:  destId,
		Type:         evm.FungibleTransfer,
		DepositNonce: nonce,
		ResourceId:   record.ResourceID,
		Payload: []interface{}{
			record.MetaData,
		},
	}, nil
}
