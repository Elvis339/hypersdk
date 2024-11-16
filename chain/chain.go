// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"context"

	"github.com/ava-labs/avalanchego/trace"
	"github.com/ava-labs/avalanchego/utils/logging"
	"github.com/ava-labs/avalanchego/x/merkledb"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/ava-labs/hypersdk/internal/workers"
	"github.com/ava-labs/hypersdk/state"
)

type Chain struct {
	builder     *Builder
	processor   *Processor
	accepter    *Accepter
	preExecutor *PreExecutor
	blockParser *BlockParser
}

func NewChain(
	tracer trace.Tracer,
	registerer *prometheus.Registry,
	parser Parser,
	mempool Mempool,
	logger logging.Logger,
	ruleFactory RuleFactory,
	metadataManager MetadataManager,
	balanceHandler BalanceHandler,
	authVerifiers workers.Workers,
	authVM AuthVM,
	validityWindow *TimeValidityWindow,
	config Config,
) (*Chain, error) {
	metrics, err := newMetrics(registerer)
	if err != nil {
		return nil, err
	}
	return &Chain{
		builder: NewBuilder(
			tracer,
			ruleFactory,
			logger,
			metadataManager,
			balanceHandler,
			mempool,
			validityWindow,
			metrics,
			config,
		),
		processor: NewProcessor(
			tracer,
			logger,
			ruleFactory,
			authVerifiers,
			authVM,
			metadataManager,
			balanceHandler,
			validityWindow,
			metrics,
			config,
		),
		accepter: NewAccepter(
			tracer,
			validityWindow,
			metrics,
		),
		preExecutor: NewPreExecutor(
			ruleFactory,
			validityWindow,
			metadataManager,
			balanceHandler,
		),
		blockParser: NewBlockParser(tracer, parser),
	}, nil
}

func (c *Chain) BuildBlock(ctx context.Context, parentView state.View, parent *ExecutionBlock) (*ExecutionBlock, *ExecutedBlock, merkledb.View, error) {
	return c.builder.BuildBlock(ctx, parentView, parent)
}

func (c *Chain) Execute(
	ctx context.Context,
	parentView state.View,
	b *ExecutionBlock,
) (*ExecutedBlock, merkledb.View, error) {
	return c.processor.Execute(ctx, parentView, b)
}

func (c *Chain) AsyncVerify(
	ctx context.Context,
	b *ExecutionBlock,
) error {
	return c.processor.AsyncVerify(ctx, b)
}

func (c *Chain) AcceptBlock(ctx context.Context, blk *ExecutionBlock) error {
	return c.accepter.AcceptBlock(ctx, blk)
}

func (c *Chain) PreExecute(
	ctx context.Context,
	parentBlk *ExecutionBlock,
	view state.View,
	tx *Transaction,
	verifyAuth bool,
) error {
	return c.preExecutor.PreExecute(ctx, parentBlk, view, tx, verifyAuth)
}

func (c *Chain) ParseBlock(ctx context.Context, bytes []byte) (*ExecutionBlock, error) {
	return c.blockParser.ParseBlock(ctx, bytes)
}