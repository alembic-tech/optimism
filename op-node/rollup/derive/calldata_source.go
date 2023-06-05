package derive

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"

	"github.com/ethereum-optimism/optimism/da"
	"github.com/ethereum-optimism/optimism/op-node/eth"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
)

type DataIter interface {
	Next(ctx context.Context) (eth.Data, error)
}

type L1TransactionFetcher interface {
	InfoAndTxsByHash(ctx context.Context, hash common.Hash) (eth.BlockInfo, types.Transactions, error)
}

// DataSourceFactory readers raw transactions from a given block & then filters for
// batch submitter transactions.
// This is not a stage in the pipeline, but a wrapper for another stage in the pipeline
type DataSourceFactory struct {
	log     log.Logger
	cfg     *rollup.Config
	fetcher L1TransactionFetcher
  da *da.Client
}

func NewDataSourceFactory(log log.Logger, cfg *rollup.Config, fetcher L1TransactionFetcher) *DataSourceFactory {
	return NewDataSourceFactoryWithDA(log, cfg, fetcher, nil)
}

func NewDataSourceFactoryWithDA(log log.Logger, cfg *rollup.Config, fetcher L1TransactionFetcher, da *da.Client) *DataSourceFactory {
  return &DataSourceFactory{log: log, cfg: cfg, fetcher: fetcher, da: da}
}

// OpenData returns a DataIter. This struct implements the `Next` function.
func (ds *DataSourceFactory) OpenData(ctx context.Context, id eth.BlockID, batcherAddr common.Address) DataIter {
	return NewDataSourceWithDA(ctx, ds.log, ds.cfg, ds.fetcher, id, batcherAddr, ds.da)
}

// DataSource is a fault tolerant approach to fetching data.
// The constructor will never fail & it will instead re-attempt the fetcher
// at a later point.
type DataSource struct {
	// Internal state + data
	open bool
	data []eth.Data
	// Required to re-attempt fetching
	id      eth.BlockID
	cfg     *rollup.Config // TODO: `DataFromEVMTransactions` should probably not take the full config
	fetcher L1TransactionFetcher
	da      *da.Client
	log     log.Logger

	batcherAddr common.Address
}

// NewDataSource creates a new calldata source. It suppresses errors in fetching the L1 block if they occur.
// If there is an error, it will attempt to fetch the result on the next call to `Next`.
func NewDataSource(ctx context.Context, log log.Logger, cfg *rollup.Config, fetcher L1TransactionFetcher, block eth.BlockID, batcherAddr common.Address) DataIter {
  return NewDataSourceWithDA(ctx, log, cfg, fetcher, block, batcherAddr, nil)
}

// NewDataSourceWithDA creates a new calldata source. It suppresses errors in fetching the L1 block if they occur.
// If there is an error, it will attempt to fetch the result on the next call to `Next`.
func NewDataSourceWithDA(ctx context.Context, log log.Logger, cfg *rollup.Config, fetcher L1TransactionFetcher, block eth.BlockID, batcherAddr common.Address, da *da.Client) DataIter {
	_, txs, err := fetcher.InfoAndTxsByHash(ctx, block.Hash)
	if err != nil {
		return &DataSource{
			open:        false,
			id:          block,
			cfg:         cfg,
			fetcher:     fetcher,
			log:         log,
			batcherAddr: batcherAddr,
      da: da,
		}
	} else {
		data, err := DataFromEVMTransactions(cfg, batcherAddr, txs, da, log.New("origin", block))
		if err != nil {
			return &DataSource{
				open:        false,
				id:          block,
				cfg:         cfg,
				fetcher:     fetcher,
				log:         log,
				batcherAddr: batcherAddr,
        da: da,
			}
		}
		return &DataSource{
			open: true,
			data: data,
		}
	}
}

// Next returns the next piece of data if it has it. If the constructor failed, this
// will attempt to reinitialize itself. If it cannot find the block it returns a ResetError
// otherwise it returns a temporary error if fetching the block returns an error.
func (ds *DataSource) Next(ctx context.Context) (eth.Data, error) {
	if !ds.open {
		if _, txs, err := ds.fetcher.InfoAndTxsByHash(ctx, ds.id.Hash); err == nil {
			ds.open = true

			data, err := DataFromEVMTransactions(ds.cfg, ds.batcherAddr, txs, ds.da, log.New("origin", ds.id))
			if err != nil {
				return nil, NewTemporaryError(fmt.Errorf("failed to retrieve data from transaction: %w", err))
			}
			ds.data = data
		} else if errors.Is(err, ethereum.NotFound) {
			return nil, NewResetError(fmt.Errorf("failed to open calldata source: %w", err))
		} else {
			return nil, NewTemporaryError(fmt.Errorf("failed to open calldata source: %w", err))
		}
	}
	if len(ds.data) == 0 {
		return nil, io.EOF
	} else {
		data := ds.data[0]
		ds.data = ds.data[1:]
		return data, nil
	}
}

// DataFromEVMTransactions filters all of the transactions and returns the calldata from transactions
// that are sent to the batch inbox address from the batch sender address.
// This will return an empty array if no valid transactions are found.
func DataFromEVMTransactions(config *rollup.Config, batcherAddr common.Address, txs types.Transactions, daClient *da.Client, log log.Logger) ([]eth.Data, error) {
	var out []eth.Data
	l1Signer := config.L1Signer()
	for j, tx := range txs {
		to := tx.To()
		if to != nil && *to == config.BatchInboxAddress {
			seqDataSubmitter, err := l1Signer.Sender(tx) // optimization: only derive sender if To is correct
			if err != nil {
				log.Warn("tx in inbox with invalid signature", "index", j, "err", err)
				continue // bad signature, ignore
			}
			// some random L1 user might have sent a transaction to our batch inbox, ignore them
			if seqDataSubmitter != batcherAddr {
				log.Warn("tx in inbox with unauthorized submitter", "index", j, "err", err)
				continue // not an authorized batch submitter, ignore
			}

			out = append(out, tx.Data())
		} else if to != nil && *to == config.DataAvailabilityInboxAddress {

      log.Info("da tx detected", "hash", tx.Hash())

			if daClient == nil {
        return nil, errors.New("alternative DA transaction found but no DA client")
			}

			seqDataSubmitter, err := l1Signer.Sender(tx) // optimization: only derive sender if To is correct
			if err != nil {
				log.Warn("da tx in inbox with invalid signature", "index", j, "err", err)
				continue // bad signature, ignore
			}
			// some random L1 user might have sent a transaction to our batch inbox, ignore them
			if seqDataSubmitter != batcherAddr {
				log.Warn("da tx in inbox with unauthorized submitter", "index", j, "err", err)
				continue // not an authorized batch submitter, ignore
			}

      if len(tx.Data()) != 33 || tx.Data()[0] != da.CentralizedBatchHeaderID {
        log.Warn("invalid DA batch header", "tx_hash", tx.Hash())
        continue
      }
      dataHash := tx.Data()[1:]
			data, err := daClient.GetBatch(dataHash)
			fmt.Println("-----------data from DA", data)
			if err != nil {
				return nil, err
			}
			out = append(out, data)
		}
	}
  log.Info("adding data from evm calldata", "len", len(out))
	return out, nil
}
