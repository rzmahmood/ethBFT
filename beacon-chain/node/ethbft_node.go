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
	executionCli "github.com/prysmaticlabs/prysm/v5/cmd/beacon-chain/execution"
	"github.com/prysmaticlabs/prysm/v5/consensus-types/blocks"
	payloadattribute "github.com/prysmaticlabs/prysm/v5/consensus-types/payload-attribute"
	"github.com/prysmaticlabs/prysm/v5/network"
	"github.com/prysmaticlabs/prysm/v5/network/authorization"
	enginev1 "github.com/prysmaticlabs/prysm/v5/proto/engine/v1"
	pb "github.com/prysmaticlabs/prysm/v5/proto/engine/v1"
	"github.com/urfave/cli/v2"
)

type BFTNode struct {
	p2p             *P2PNode
	executionClient *rpc.Client
	isSequencer     bool
}

// NewBFTNode creates a new node instance, sets up configuration options, and registers
// every required service to the node.
func NewBFTNode(cliCtx *cli.Context, cancel context.CancelFunc, opts ...Option) (*BFTNode, error) {
	ctx := cliCtx.Context

	// Create connection to the Execution Client
	jwtSecret, err := executionCli.ParseJWTSecretFromFile(cliCtx)
	if err != nil {
		return nil, errors.Wrap(err, "could not read JWT secret file for authenticating execution API")
	}
	authExecutionEndpoint, err := executionCli.ParseExecutionChainEndpoint(cliCtx)
	if err != nil {
		return nil, err
	}

	isSequencer := false
	// hacky way to distinguish validator from RPC
	if authExecutionEndpoint == "http://localhost:8200" {
		fmt.Println("This node is the sequencer")
		isSequencer = true
	} else {
		fmt.Println("This node is the rpc node")
	}

	fmt.Println("ETHBFT: NewExecutionRPCClient")
	hEndpoint := network.HttpEndpoint(authExecutionEndpoint)
	hEndpoint.Auth.Method = authorization.Bearer
	hEndpoint.Auth.Value = string(jwtSecret)
	executionClientRPC, err := network.NewExecutionRPCClient(ctx, hEndpoint, nil)
	if err != nil {
		return nil, err
	}

	// Set up the P2P Service
	fmt.Println("ETHBFT: Creating P2P Service")
	p2pNode, err := startP2P(ctx, cliCtx, isSequencer)
	if err != nil {
		return nil, err
	}

	return &BFTNode{p2p: p2pNode, executionClient: executionClientRPC, isSequencer: isSequencer}, nil
}

func (b *BFTNode) rpcLoop(ctx context.Context) error {
	blockTicker := time.NewTicker(2 * time.Second)
	defer blockTicker.Stop()
	for _ = range blockTicker.C {
		fmt.Println("RPC Loop Peers: ", b.p2p.Host.Peerstore().Peers())
	}
	return nil
}

func (b *BFTNode) validatorLoop(ctx context.Context) error {
	//fmt.Println("Blocking Validator Loop until peer discovered")
	//newPeer := <-b.p2p.PeerChan
	//fmt.Println("Peer discovered in validator!: ", newPeer.String())

	blockTicker := time.NewTicker(2 * time.Second)
	defer blockTicker.Stop()

	// Get the block to start from in the execution client
	latestBlock, err := ethclient.NewClient(b.executionClient).BlockByNumber(ctx, nil)
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
		err = b.executionClient.CallContext(ctx, result, "engine_forkchoiceUpdatedV3", f, a)
		if err != nil {
			return err
		}
		fmt.Println("ETHBFT: Executed fork choice update, with status", result.Status.String())

		// Now we have initial block state setup
		// We call 	payload, err := vs.ExecutionEngineCaller.GetPayload(ctx, *payloadID, slot)
		getPayloadResult := &pb.ExecutionPayloadDenebWithValueAndBlobsBundle{}
		err = b.executionClient.CallContext(ctx, getPayloadResult, "engine_getPayloadV3", result.PayloadId)
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
		err = b.executionClient.CallContext(ctx, newPayloadResult, "engine_newPayloadV3", payloadPb, []common.Hash{}, common.BytesToHash(payloadPb.ParentHash))
		if err != nil {
			return err
		}
		fmt.Printf("ETHBFT: [Block Number: %d] Called engine_newPayloadV3: %s\n", payloadPb.BlockNumber, newPayloadResult.String())
		hashToFinalize = payloadPb.BlockHash

		// Marshal so it can be sent over the wire
		marshalledBlock, err := payloadPb.MarshalSSZ()
		if err != nil {
			return err
		}

		if err := b.p2p.Topic.Publish(ctx, marshalledBlock); err != nil {
			return err
		}

		fmt.Printf("Published block %d\n", payloadPb.BlockNumber)
	}
	return nil
}

func (b *BFTNode) listenOnSubscription(ctx context.Context) error {

	go func() {
		for {
			fmt.Printf("[SUB] Attempting to receive message from topic\n")
			msg, err := b.p2p.Subscription.Next(ctx)
			if err != nil {
				fmt.Printf("Failed to receive message: %v\n", err)
				continue
			}
			fmt.Println("[SUB] Message received, attempting to unmarshal")

			var block pb.ExecutionPayloadDeneb
			err = block.UnmarshalSSZ(msg.Data) // Assuming the message data is in SSZ format
			if err != nil {
				fmt.Printf("[SUB] failed to unmarshal block: %v\n", err)
				continue
			}

			// Here you can process the block as needed
			fmt.Printf("[SUB] Received new block with hash: %s and number: %d from %s \n", hex.EncodeToString(block.BlockHash), block.BlockNumber, msg.ReceivedFrom.String())
		}
	}()
	return nil
}

func (b *BFTNode) Start() error {
	// Start P2p Service
	ctx := context.Background()

	// Subscribe to block topic
	if err := b.listenOnSubscription(ctx); err != nil {
		return fmt.Errorf("failed to setup block subscription: %v", err)
	}

	if b.isSequencer {
		if err := b.validatorLoop(ctx); err != nil {
			return err
		}
	} else {
		if err := b.rpcLoop(ctx); err != nil {
			return err
		}
	}
	return nil
}
