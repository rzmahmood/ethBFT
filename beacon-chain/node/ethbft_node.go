// Package node is the main service which launches a beacon node and manages
// the lifecycle of all its associated services at runtime, such as p2p, RPC, sync,
// gracefully closing them if the process ends.
package node

import (
	"context"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/pkg/errors"
	"github.com/prysmaticlabs/prysm/v5/beacon-chain/execution"
	"github.com/prysmaticlabs/prysm/v5/beacon-chain/node/registration"
	"github.com/prysmaticlabs/prysm/v5/beacon-chain/p2p"
	"github.com/prysmaticlabs/prysm/v5/cmd"
	executionCli "github.com/prysmaticlabs/prysm/v5/cmd/beacon-chain/execution"
	"github.com/prysmaticlabs/prysm/v5/consensus-types/blocks"
	payloadattribute "github.com/prysmaticlabs/prysm/v5/consensus-types/payload-attribute"
	"github.com/prysmaticlabs/prysm/v5/container/slice"
	"github.com/prysmaticlabs/prysm/v5/network"
	"github.com/prysmaticlabs/prysm/v5/network/authorization"
	enginev1 "github.com/prysmaticlabs/prysm/v5/proto/engine/v1"
	pb "github.com/prysmaticlabs/prysm/v5/proto/engine/v1"
	"github.com/urfave/cli/v2"
)

type BFTNode struct {
	p2p             *p2p.Service
	engineAPIClient *rpc.Client
	executionClient *ethclient.Client
}

func createP2P(ctx context.Context, cliCtx *cli.Context) (*p2p.Service, error) {
	bootstrapNodeAddrs, dataDir, err := registration.P2PPreregistration(cliCtx)
	if err != nil {
		return nil, errors.Wrapf(err, "could not register p2p service")
	}

	svc, err := p2p.NewService(ctx, &p2p.Config{
		NoDiscovery:          cliCtx.Bool(cmd.NoDiscovery.Name),
		StaticPeers:          slice.SplitCommaSeparated(cliCtx.StringSlice(cmd.StaticPeers.Name)),
		Discv5BootStrapAddrs: p2p.ParseBootStrapAddrs(bootstrapNodeAddrs),
		RelayNodeAddr:        cliCtx.String(cmd.RelayNode.Name),
		DataDir:              dataDir,
		LocalIP:              cliCtx.String(cmd.P2PIP.Name),
		HostAddress:          cliCtx.String(cmd.P2PHost.Name),
		HostDNS:              cliCtx.String(cmd.P2PHostDNS.Name),
		PrivateKey:           cliCtx.String(cmd.P2PPrivKey.Name),
		StaticPeerID:         cliCtx.Bool(cmd.P2PStaticID.Name),
		MetaDataDir:          cliCtx.String(cmd.P2PMetadata.Name),
		QUICPort:             cliCtx.Uint(cmd.P2PQUICPort.Name),
		TCPPort:              cliCtx.Uint(cmd.P2PTCPPort.Name),
		UDPPort:              cliCtx.Uint(cmd.P2PUDPPort.Name),
		MaxPeers:             cliCtx.Uint(cmd.P2PMaxPeers.Name),
		QueueSize:            cliCtx.Uint(cmd.PubsubQueueSize.Name),
		AllowListCIDR:        cliCtx.String(cmd.P2PAllowList.Name),
		DenyListCIDR:         slice.SplitCommaSeparated(cliCtx.StringSlice(cmd.P2PDenyList.Name)),
		EnableUPnP:           cliCtx.Bool(cmd.EnableUPnPFlag.Name),
		//StateNotifier:        b,
		//DB:                   b.db,
		//ClockWaiter:          b.clockWaiter,
	})
	if err != nil {
		return nil, err
	}
	return svc, nil
}

// NewBFTNode creates a new node instance, sets up configuration options, and registers
// every required service to the node.
func NewBFTNode(cliCtx *cli.Context, cancel context.CancelFunc, opts ...Option) (*BFTNode, error) {
	ctx := cliCtx.Context

	nodeType := cliCtx.String(cmd.VerbosityFlag.Value)
	fmt.Printf("ETHBFT: Nodetype is %s\n", nodeType)

	// Create connection to the Execution Client
	jwtSecret, err := executionCli.ParseJWTSecretFromFile(cliCtx)
	if err != nil {
		return nil, errors.Wrap(err, "could not read JWT secret file for authenticating execution API")
	}
	authExecutionEndpoint, err := executionCli.ParseExecutionChainEndpoint(cliCtx)
	if err != nil {
		return nil, err
	}
	fmt.Println("ETHBFT: NewExecutionRPCClient")
	hEndpoint := network.HttpEndpoint(authExecutionEndpoint)
	hEndpoint.Auth.Method = authorization.Bearer
	hEndpoint.Auth.Value = string(jwtSecret)
	engineAPIClient, err := network.NewExecutionRPCClient(ctx, hEndpoint, nil)
	if err != nil {
		return nil, err
	}
	
	execClient, err := ethclient.DialContext(ctx, "http://localhost:8000")
	if err != nil {
		return nil, err
	}

	// Set up the P2P Service
	fmt.Println("ETHBFT: Creating P2P Service")
	p2pService, err := createP2P(ctx, cliCtx)
	if err != nil {
		return nil, err
	}

	return &BFTNode{p2p: p2pService, executionClient: execClient, engineAPIClient: engineAPIClient}, nil
}

func (b *BFTNode) Start() error {
	// Start P2p Service
	ctx := context.Background()

	fmt.Println("Starting P2P Service")
	go b.p2p.Start()
	// Get ENR
	time.Sleep(5 * time.Second)
	serializedEnr, err := p2p.SerializeENR(b.p2p.ENR())
	if err != nil {
		return err
	}

	fmt.Printf("The serialized ENR is: %s\n", serializedEnr)

	blockTicker := time.NewTicker(2 * time.Second)
	defer blockTicker.Stop()

	// Get the block to start from in the execution client
	latestBlock, err := b.executionClient.BlockByNumber(ctx, nil)
	if err != nil {
		return err
	}
	hashToFinalize := latestBlock.Hash().Bytes()
	for _ = range blockTicker.C {
		fmt.Printf("ETHBFT: Latest blockhash is %s\n", hex.EncodeToString(hashToFinalize))
		timeStamp := uint64(time.Now().Unix())
		f := &enginev1.ForkchoiceState{
			HeadBlockHash:      hashToFinalize,
			SafeBlockHash:      hashToFinalize,
			FinalizedBlockHash: hashToFinalize,
		}
		// Arbitrary feeRecipient and randao
		feeRecipient := common.HexToAddress("0xFe8664457176D0f87EAaBd103ABa410855F81010")
		randao := common.HexToHash("0xc48549953ec32ef7cacfd9812de1290bab71de5e5a08d4ea4383d6a2d3754a7c")
		// empty withdrawals
		withdrawals := make([]*enginev1.Withdrawal, 0, 0)
		attr, err := payloadattribute.New(&enginev1.PayloadAttributesV3{
			// unsafe cast
			Timestamp:             timeStamp,
			PrevRandao:            randao.Bytes(),
			SuggestedFeeRecipient: feeRecipient.Bytes(),
			Withdrawals:           withdrawals,
			ParentBeaconBlockRoot: hashToFinalize,
		})
		if err != nil {
			return err
		}
		a, err := attr.PbV3()
		if err != nil {
			return err
		}
		fmt.Println("Attempting to call FCUV3")
		result := &execution.ForkchoiceUpdatedResponse{}
		err = b.engineAPIClient.CallContext(ctx, result, "engine_forkchoiceUpdatedV3", f, a)
		if err != nil {
			return err
		}
		fmt.Println("ETHBFT: Executed fork choice update, with status", result.Status.String())

		// Now we have initial block state setup
		// We call 	payload, err := vs.ExecutionEngineCaller.GetPayload(ctx, *payloadID, slot)
		getPayloadResult := &pb.ExecutionPayloadDenebWithValueAndBlobsBundle{}
		err = b.engineAPIClient.CallContext(ctx, getPayloadResult, "engine_getPayloadV3", result.PayloadId)
		if err != nil {
			fmt.Println("Failed to called getPayloadV3", err.Error())
			return err
		}
		fmt.Println("ETHBFT: Called engine_getPayloadV3")
		execPayloadWrapped, err := blocks.WrappedExecutionPayloadDeneb(getPayloadResult.Payload, blocks.PayloadValueToWei(getPayloadResult.Value))
		if err != nil {
			fmt.Println("Failed to WrappedExecutionPayloadDeneb", err.Error())
			return err
		}

		// Now we have the payload, we call NewPayloadV2
		// Now that we have the execution payload, we need to tell the execution client that a new possible block exists
		fmt.Println("ETHBFT: Called Wrapped Execution Payload")
		payloadPb, ok := execPayloadWrapped.Proto().(*pb.ExecutionPayloadDeneb)
		if !ok {
			return errors.New("execution data must be a Deneb execution payload")
		}

		newPayloadResult := &pb.PayloadStatus{}
		// https://github.com/ethereum/consensus-specs/blob/dev/specs/deneb/beacon-chain.md#beacon-chain-state-transition-function
		fmt.Printf("ETHBFT: Calling newPayload with body hash: %s\n", common.BytesToHash(payloadPb.BlockHash).String())
		err = b.engineAPIClient.CallContext(ctx, newPayloadResult, "engine_newPayloadV3", payloadPb, []common.Hash{}, common.BytesToHash(payloadPb.ParentHash))
		if err != nil {
			return err
		}
		fmt.Printf("ETHBFT: [Block Number: %d] Called engine_newPayloadV3: %s\n", payloadPb.BlockNumber, newPayloadResult.String())
		hashToFinalize = payloadPb.BlockHash
	}
	return nil
}
