package gateway

import (
	"github.com/pkg/errors"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel/invoke"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
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
