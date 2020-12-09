/*
Copyright 2020 Datachain All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package gateway

import (
	"github.com/pkg/errors"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel/invoke"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/event"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
)

// Execute a transaction to the ledger. The transaction function represented by this object
// will be evaluated on the endorsing peers and then submitted to the ordering service
// for committing to the ledger.
func (txn *Transaction) Execute(args ...string) (*channel.Response, error) {
	bytes := make([][]byte, len(args))
	for i, v := range args {
		bytes[i] = []byte(v)
	}
	txn.request.Args = bytes

	var options []channel.RequestOption
	if txn.endorsingPeers != nil {
		options = append(options, channel.WithTargetEndpoints(txn.endorsingPeers...))
	}
	options = append(options, channel.WithTimeout(fab.Execute, txn.contract.network.gateway.options.Timeout))

	response, err := txn.contract.client.InvokeHandler(
		newSubmitHandler(txn.eventch),
		*txn.request,
		options...,
	)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to submit")
	}

	return &response, nil
}

// Simulate a transaction to the ledger.
// The transaction function represented by this object
// will be evaluated on the endorsing peers.
// but transaction is not sent to the ordering service
func (txn *Transaction) Simulate(args ...string) (*channel.Response, error) {
	bytes := make([][]byte, len(args))
	for i, v := range args {
		bytes[i] = []byte(v)
	}
	txn.request.Args = bytes

	var options []channel.RequestOption
	if txn.endorsingPeers != nil {
		options = append(options, channel.WithTargetEndpoints(txn.endorsingPeers...))
	}
	options = append(options, channel.WithTimeout(fab.Execute, txn.contract.network.gateway.options.Timeout))

	response, err := txn.contract.client.InvokeHandler(
		invoke.NewSelectAndEndorseHandler(
			invoke.NewEndorsementValidationHandler(
				invoke.NewSignatureValidationHandler(),
			),
		),
		*txn.request,
		options...,
	)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to submit")
	}

	return &response, nil
}

// GetNetworkWithEventOption returns an object representing a network channel.
//  Parameters:
//  name is the name of the network channel
//  eventOpts is the event client options for the event service of the network
//
//  Returns:
//  A Network object representing the channel
func (gw *Gateway) GetNetworkWithEventOption(name string, eventOpts ...event.ClientOption) (*Network, error) {
	var channelProvider context.ChannelProvider
	if gw.options.Identity != nil {
		channelProvider = gw.sdk.ChannelContext(name, fabsdk.WithIdentity(gw.options.Identity), fabsdk.WithOrg(gw.org))
	} else {
		channelProvider = gw.sdk.ChannelContext(name, fabsdk.WithUser(gw.options.User), fabsdk.WithOrg(gw.org))
	}
	return newNetworkWithEventOption(gw, channelProvider, eventOpts...)
}

// newNetworkWithEventOption returns an object representing a network channel with the explicit event client options.
func newNetworkWithEventOption(gateway *Gateway, channelProvider context.ChannelProvider, eventOpts ...event.ClientOption) (*Network, error) {
	n := Network{
		gateway: gateway,
	}

	// Channel client is used to query and execute transactions
	client, err := channel.New(channelProvider)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create new channel client")
	}

	n.client = client

	ctx, err := channelProvider()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create new channel context")
	}

	n.name = ctx.ChannelID()

	n.event, err = event.New(channelProvider, eventOpts...)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create new event client")
	}

	return &n, nil
}
