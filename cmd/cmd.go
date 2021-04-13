package cmd

import (
	"math/big"

	"github.com/ChainSafe/chainbridgev2/relayer"

	"github.com/ChainSafe/chainbridge-utils/keystore"
	"github.com/ChainSafe/chainbridgev2/chains/evmd"
	"github.com/ethereum/go-ethereum/common"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

var AliceKp = keystore.TestKeyRing.EthereumKeys[keystore.AliceKey]
var BobKp = keystore.TestKeyRing.EthereumKeys[keystore.BobKey]
var EveKp = keystore.TestKeyRing.EthereumKeys[keystore.EveKey]

var (
	DefaultRelayerAddresses = []common.Address{
		common.HexToAddress(keystore.TestKeyRing.EthereumKeys[keystore.AliceKey].Address()),
		common.HexToAddress(keystore.TestKeyRing.EthereumKeys[keystore.BobKey].Address()),
		common.HexToAddress(keystore.TestKeyRing.EthereumKeys[keystore.CharlieKey].Address()),
		common.HexToAddress(keystore.TestKeyRing.EthereumKeys[keystore.DaveKey].Address()),
		common.HexToAddress(keystore.TestKeyRing.EthereumKeys[keystore.EveKey].Address()),
	}
)

const DefaultGasLimit = 6721975
const DefaultGasPrice = 20000000000

const TestEndpoint = "ws://localhost:8545"
const TestEndpoint2 = "ws://localhost:8546"

func Run(ctx *cli.Context) error {
	//
	errChn := make(chan error)
	stopChn := make(chan struct{})
	//
	c, err := evmd.NewClient(TestEndpoint, false, AliceKp, big.NewInt(DefaultGasLimit), big.NewInt(DefaultGasPrice), stopChn, errChn)
	if err != nil {
		panic(err)
	}
	celoC, err := evmd.NewClient(TestEndpoint2, false, AliceKp, big.NewInt(DefaultGasLimit), big.NewInt(DefaultGasPrice), stopChn, errChn)
	if err != nil {
		panic(err)
	}
	bridgeAddress := ethcommon.HexToAddress("0x62877dDCd49aD22f5eDfc6ac108e9a4b5D2bD88B")
	ethListener := evmd.NewListener(c, bridgeAddress, 1)
	celoListener := evmd.NewListener(celoC, bridgeAddress, 2)

	ethListener.RegisterHandler(common.HexToAddress("0x3167776db165D8eA0f51790CA2bbf44Db5105ADF"), evmd.HandleErc20DepositedEvent)
	celoListener.RegisterHandler(common.HexToAddress("0x3167776db165D8eA0f51790CA2bbf44Db5105ADF"), evmd.HandleErc20DepositedEvent)
	//
	//// It should listen different chains and accept different configs
	r := relayer.NewRelayer([]relayer.IListener{ethListener, celoListener})
	//
	//ethWriter := relayer.NewWriter(propExecuterETH)
	//celoWriter := relayer.NewWriter(propExecuterCelo)
	//
	//r.RegisterWriter(1, ethWriter)
	//r.RegisterWriter(2, celoWriter)
	//
	//go r.Start(stopChn, errChn)
	//
	//sysErr := make(chan os.Signal, 1)
	//signal.Notify(sysErr,
	//	syscall.SIGTERM,
	//	syscall.SIGINT,
	//	syscall.SIGHUP,
	//	syscall.SIGQUIT)

	select {
	case err := <-errChn:
		log.Error().Err(err).Msg("failed to listen and serve")
		close(stopChn)
		return err
	case sig := <-sysErr:
		log.Info().Msgf("terminating got [%v] signal", sig)
		return nil
	}
}
