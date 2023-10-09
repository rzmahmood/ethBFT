// Package node is the main service which launches a beacon node and manages
// the lifecycle of all its associated services at runtime, such as p2p, RPC, sync,
// gracefully closing them if the process ends.
package node

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"github.com/prysmaticlabs/prysm/v4/async/event"
	"github.com/prysmaticlabs/prysm/v4/beacon-chain/cache"
	"github.com/prysmaticlabs/prysm/v4/beacon-chain/cache/depositcache"
	"github.com/prysmaticlabs/prysm/v4/beacon-chain/db"
	"github.com/prysmaticlabs/prysm/v4/beacon-chain/db/kv"
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
	"github.com/prysmaticlabs/prysm/v4/cmd/beacon-chain/flags"
	"github.com/prysmaticlabs/prysm/v4/config/params"
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
func  NewBFTNode(cliCtx *cli.Context, opts ...Option) (*BFTNode, error) {
	log.Info("ETHBFT: Doing Nothing with New Node")
	statePath := "network/node-0/consensus/beacondata/beaconchaindata"
	if err := configureTracing(cliCtx); err != nil {
		return nil, err
	}
	ctx := cliCtx.Context


	// Configure forks related configs
	flags.ConfigureGlobalFlags(cliCtx)
	if err := configureChainConfig(cliCtx); err != nil {
		return nil, err
	}

	fmt.Println("ETHBFT: Running InitializeForkSchedule")
	params.BeaconConfig().InitializeForkSchedule()
	// End various configs

	store, err := kv.NewKVStore(ctx, statePath)
	if err != nil {
		return nil, err
	}
	if err = store.RunMigrations(ctx); err != nil {
		return nil, err
	}
	fmt.Println("ETHBFT: Running NewFileInitializer")
	genesisInitializer, err := genesis.NewFileInitializer("network/node-0/consensus/genesis.ssz")
	if err != nil {
		return nil, errors.Wrap(err, "error preparing to initialize genesis db state from local ssz files")
	}
	fmt.Println("ETHBFT: Running Initialize")
	err = genesisInitializer.Initialize(ctx, store)
	if err != nil {
		return nil, err
	}
	fmt.Println("ETHBFT: Running EnsureEmbeddedGenesis")
	err = store.EnsureEmbeddedGenesis(ctx)
	if err != nil {
		return nil, err
	}
	fmt.Println("ETHBFT: Running GenesisBlock")
	genesisBlock, err := store.GenesisBlock(ctx)
	if err != nil {
		return nil, err
	}
	var r [32]byte
	if genesisBlock != nil && !genesisBlock.IsNil() {
		r, err = genesisBlock.Block().HashTreeRoot()
		if err != nil {
			return nil, err
		}
	}
	var bs state.BeaconState
	fmt.Println("ETHBFT: Running HasState")
	if store.HasState(ctx, r) {
		bs, err = store.State(ctx, r)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("failed to load state")
	}

	fmt.Println("Successfully loaded beacon state ", bs.Version())


	// Proposing Block to Execution Engine
	jwtSecret, err := execution2.ParseJWTSecretFromFile(cliCtx)
	if err != nil {
		return nil, errors.Wrap(err, "could not read JWT secret file for authenticating execution API")
	}
	endpoint := "http://localhost:8200"
	fmt.Println("ETHBFT: NewExecutionRPCClient")
	hEndpoint := network.HttpEndpoint(endpoint)
	hEndpoint.Auth.Method = authorization.Bearer
	hEndpoint.Auth.Value = string(jwtSecret)

	client, err := network.NewExecutionRPCClient(ctx, hEndpoint)
	if err != nil {
		return nil, err
	}
	fmt.Println("ETHBFT: Getting latest execution payload header")
	header, err := bs.LatestExecutionPayloadHeader()
	if err != nil {
		return nil, err
	}

	blockHash := header.BlockHash()
	timeStamp := uint64(time.Now().Unix())
	startTime := time.Now()
	numBlocks := 10000

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
			Timestamp: timeStamp,
			PrevRandao:            randao.Bytes(),
			SuggestedFeeRecipient: feeRecipient.Bytes(),
			Withdrawals: withdrawals,
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
		newPayloadResult := &pb.PayloadStatus{}
		err = client.CallContext(ctx, newPayloadResult, "engine_newPayloadV2", payloadPb)
		if err != nil {
			return nil, err
		}
		fmt.Printf("ETHBFT: [Block Number: %d] Called engine_newPayloadV2: %s\n", payloadPb.BlockNumber, newPayloadResult.String())
		blockHash = payloadPb.BlockHash
		timeStamp++
	}
	timeTaken := time.Since(startTime)
	fmt.Printf("%d blocks took %s\n", numBlocks, timeTaken.String())
	return nil, nil
}

func (b *BFTNode) Start() {
	log.Info("ETHBFT: Doing nothing in Start")
}