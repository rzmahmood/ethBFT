// Package node is the main service which launches a beacon node and manages
// the lifecycle of all its associated services at runtime, such as p2p, RPC, sync,
// gracefully closing them if the process ends.
package node

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pkg/errors"
	"github.com/prysmaticlabs/prysm/v4/async/event"
	"github.com/prysmaticlabs/prysm/v4/beacon-chain/cache"
	"github.com/prysmaticlabs/prysm/v4/beacon-chain/cache/depositcache"
	"github.com/prysmaticlabs/prysm/v4/beacon-chain/db"
	"github.com/prysmaticlabs/prysm/v4/beacon-chain/execution"
	"github.com/prysmaticlabs/prysm/v4/beacon-chain/forkchoice"
	"github.com/prysmaticlabs/prysm/v4/beacon-chain/operations/attestations"
	"github.com/prysmaticlabs/prysm/v4/beacon-chain/operations/blstoexec"
	"github.com/prysmaticlabs/prysm/v4/beacon-chain/operations/slashings"
	"github.com/prysmaticlabs/prysm/v4/beacon-chain/operations/synccommittee"
	"github.com/prysmaticlabs/prysm/v4/beacon-chain/operations/voluntaryexits"
	"github.com/prysmaticlabs/prysm/v4/beacon-chain/startup"
	"github.com/prysmaticlabs/prysm/v4/beacon-chain/state"
	"github.com/prysmaticlabs/prysm/v4/beacon-chain/state/stategen"
	"github.com/prysmaticlabs/prysm/v4/beacon-chain/sync/checkpoint"
	"github.com/prysmaticlabs/prysm/v4/beacon-chain/sync/genesis"
	execution2 "github.com/prysmaticlabs/prysm/v4/cmd/beacon-chain/execution"
	"github.com/prysmaticlabs/prysm/v4/consensus-types/blocks"
	payloadattribute "github.com/prysmaticlabs/prysm/v4/consensus-types/payload-attribute"
	"github.com/prysmaticlabs/prysm/v4/encoding/bytesutil"
	"github.com/prysmaticlabs/prysm/v4/math"
	"github.com/prysmaticlabs/prysm/v4/network"
	"github.com/prysmaticlabs/prysm/v4/network/authorization"
	enginev1 "github.com/prysmaticlabs/prysm/v4/proto/engine/v1"
	pb "github.com/prysmaticlabs/prysm/v4/proto/engine/v1"
	"github.com/prysmaticlabs/prysm/v4/runtime"
	"github.com/prysmaticlabs/prysm/v4/time"
	"github.com/urfave/cli/v2"
	"math/big"
	"sync"
	time2 "time"
)

type BFTNode struct {
	cliCtx                  *cli.Context
	ctx                     context.Context
	cancel                  context.CancelFunc
	services                *runtime.ServiceRegistry
	lock                    sync.RWMutex
	stop                    chan struct{} // Channel to wait for termination notifications.
	db                      db.Database
	slasherDB               db.SlasherDatabase
	attestationPool         attestations.Pool
	exitPool                voluntaryexits.PoolManager
	slashingsPool           slashings.PoolManager
	syncCommitteePool       synccommittee.Pool
	blsToExecPool           blstoexec.PoolManager
	depositCache            *depositcache.DepositCache
	proposerIdsCache        *cache.ProposerPayloadIDsCache
	stateFeed               *event.Feed
	blockFeed               *event.Feed
	opFeed                  *event.Feed
	stateGen                *stategen.State
	collector               *bcnodeCollector
	slasherBlockHeadersFeed *event.Feed
	slasherAttestationsFeed *event.Feed
	finalizedStateAtStartUp state.BeaconState
	serviceFlagOpts         *serviceFlagOpts
	GenesisInitializer      genesis.Initializer
	CheckpointInitializer   checkpoint.Initializer
	forkChoicer             forkchoice.ForkChoicer
	clockWaiter             startup.ClockWaiter
	initialSyncComplete     chan struct{}
}

// NewBFTNode creates a new node instance, sets up configuration options, and registers
// every required service to the node.
func NewBFTNode(cliCtx *cli.Context, opts ...Option) (*BFTNode, error) {
	ctx := cliCtx.Context

	//// Proposing Block to Execution Engine
	jwtSecret, err := execution2.ParseJWTSecretFromFile(cliCtx)
	if err != nil {
		return nil, errors.Wrap(err, "could not read JWT secret file for authenticating execution API")
	}
	endpoint := "http://localhost:8200"
	fmt.Println("ETHBFT: NewExecutionRPCClient")
	hEndpoint := network.HttpEndpoint(endpoint)
	hEndpoint.Auth.Method = authorization.Bearer
	hEndpoint.Auth.Value = string(jwtSecret)
	//
	client, err := network.NewExecutionRPCClient(ctx, hEndpoint)
	if err != nil {
		return nil, err
	}

	execClient, err := ethclient.DialContext(ctx, "http://localhost:8000")
	if err != nil {
		return nil, err
	}

	latestBlock, err := execClient.BlockByNumber(ctx, nil)
	if err != nil {
		return nil, err
	}

	//blockHash := header.BlockHash()
	timeStamp := uint64(time.Now().Unix())
	startTime := time.Now()
	numBlocks := 100
	blockHash := latestBlock.Hash().Bytes()

	for i := 0; i < numBlocks; i++ {
		fmt.Printf("ETHBFT: Latest blockhash is %s\n", hex.EncodeToString(blockHash))
		f := &enginev1.ForkchoiceState{
			HeadBlockHash:      blockHash,
			SafeBlockHash:      blockHash,
			FinalizedBlockHash: blockHash,
		}
		feeRecipient := common.HexToAddress("0xFe8664457176D0f87EAaBd103ABa410855F81010")
		randao := common.HexToHash("0xc48549953ec32ef7cacfd9812de1290bab71de5e5a08d4ea4383d6a2d3754a7c")
		withdrawals := make([]*enginev1.Withdrawal, 0, 0)
		fmt.Println("ETHBFT: Using timestamp: ", timeStamp)
		attr, err := payloadattribute.New(&enginev1.PayloadAttributesV2{
			// unsafe cast
			Timestamp:             timeStamp,
			PrevRandao:            randao.Bytes(),
			SuggestedFeeRecipient: feeRecipient.Bytes(),
			Withdrawals:           withdrawals,
		})
		a, err := attr.PbV2()
		if err != nil {
			return nil, err
		}
		result := &execution.ForkchoiceUpdatedResponse{}
		err = client.CallContext(ctx, result, "engine_forkchoiceUpdatedV2", f, a)
		if err != nil {
			return nil, err
		}
		fmt.Println("ETHBFT: Executed for choice update, with status", result.Status.String())

		// Now we have initial block state setup
		// We call 	payload, err := vs.ExecutionEngineCaller.GetPayload(ctx, *payloadID, slot)
		getPayloadResult := &pb.ExecutionPayloadCapellaWithValue{}
		err = client.CallContext(ctx, getPayloadResult, "engine_getPayloadV2", result.PayloadId)
		if err != nil {
			return nil, err
		}
		v := big.NewInt(0).SetBytes(bytesutil.ReverseByteOrder(getPayloadResult.Value))
		execPayloadWrapped, err := blocks.WrappedExecutionPayloadCapella(getPayloadResult.Payload, math.WeiToGwei(v))
		if err != nil {
			return nil, err
		}

		// Now we have the payload, we call NewPayloadV2
		// Now that we have the execution payload, we need to tell the execution client that a new possible block exists
		payloadPb, ok := execPayloadWrapped.Proto().(*pb.ExecutionPayloadCapella)
		if !ok {
			return nil, errors.New("execution data must be a Capella execution payload")
		}

		// DO

		newPayloadResult := &pb.PayloadStatus{}
		err = client.CallContext(ctx, newPayloadResult, "engine_newPayloadV2", payloadPb)
		if err != nil {
			return nil, err
		}
		fmt.Printf("ETHBFT: [Block Number: %d] Called engine_newPayloadV2: %s\n", payloadPb.BlockNumber, newPayloadResult.String())
		blockHash = payloadPb.BlockHash
		timeStamp++
		fmt.Println("Sleeping for 2 seconds.")
		time2.Sleep(2 * time2.Second)
	}
	timeTaken := time.Since(startTime)
	fmt.Printf("%d blocks took %s\n", numBlocks, timeTaken.String())
	return nil, nil
}

func (b *BFTNode) Start() {
	log.Info("ETHBFT: Doing nothing in Start")
}
