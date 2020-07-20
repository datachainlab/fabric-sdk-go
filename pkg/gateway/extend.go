package gateway

import (
	"github.com/pkg/errors"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel/invoke"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
)

// Query a transaction function and return its results.
// The transaction function will be evaluated on the endorsing peers but
// the responses will not be sent to the ordering service and hence will
// not be committed to the ledger. This can be used for querying the world state.
func (txn *Transaction) Query(args ...string) (*channel.Response, error) {
	bytes := make([][]byte, len(args))
	for i, v := range args {
		bytes[i] = []byte(v)
	}
	txn.request.Args = bytes

	var options []channel.RequestOption
	options = append(options, channel.WithTimeout(fab.Query, txn.contract.network.gateway.options.Timeout))

	response, err := txn.contract.client.Query(
		*txn.request,
		options...,
	)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to evaluate")
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
