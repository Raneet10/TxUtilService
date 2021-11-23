package txutil

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	log "github.com/inconshreveable/log15"
)

var ErrorTimeOut = errors.New("Time out")

type TxUtilService struct {
	privateKey *ecdsa.PrivateKey
	ethclient  *ethclient.Client
	GasInfo    GasInfo

	Txreceipt map[common.Hash]uint64 // TODO(raneet10): Use a DB instead?
	Txchan    chan *types.Transaction

	ctx    context.Context
	logger log.Logger // TODO(raneet10): Might be better to write own logger
}

// TODO(raneet10): Move this somewhere
type GasInfo struct {
	safeLow  int64
	standard int64
	fast     int64
	fastest  int64
}

func NewTxUtilService(pk *ecdsa.PrivateKey, ethclient *ethclient.Client, ctx context.Context, logger log.Logger) *TxUtilService {
	return &TxUtilService{
		privateKey: pk,
		ctx:        ctx,
		ethclient:  ethclient,
		Txreceipt:  make(map[common.Hash]uint64),
		Txchan:     make(chan *types.Transaction),
		logger:     logger,
	}
}

func (ts *TxUtilService) HandleTransactions(ctx context.Context, tx *types.Transaction) error {
	ts.logger.Debug("transaction received!", tx.Hash())
	newtx, err := ts.sendTransaction(ctx, tx)
	if err != nil {
		return err
	}

	ts.logger.Debug("sent tx:", newtx.Hash())

	//	ts.logger.Debug("checking tx receipt...")

	/*receipt, err := tc.checkTransactionReceipt(ctx, newtx)
	if err != nil {
		if err == ErrorTimeOut {
			tc.sendTransaction(ctx, newtx)
		} else {
			return err
		}
	}
	if receipt.Status == types.ReceiptStatusSuccessful || receipt.Status == types.ReceiptStatusFailed {
		// TODO(raneet10): check for block confirmations
		//tc.logger.Debug("Transaction receipt", receipt.Status)
		fmt.Println("Transaction receipt", receipt.Status)
		tc.Txreceipt[tx.Hash()] = receipt.Status
		return nil
	}*/

	return nil
}

func (ts *TxUtilService) sendTransaction(ctx context.Context, tx *types.Transaction) (*types.Transaction, error) {

	// calculate current gas fee
	err := ts.GetGasfee()
	if err != nil {
		return nil, err
	}
	//tc.logger.Debug("Sending transcation!", tx.Hash())

	// send transcation with the fastest value and same nonce
	newtx := tx
	chainId, err := ts.ethclient.NetworkID(ctx)
	if err != nil {
		return nil, err
	}

	newtx.GasPrice().SetInt64(ts.GasInfo.fast)
	sigTx, err := types.SignTx(newtx, types.NewEIP155Signer(chainId), ts.privateKey)
	ts.logger.Debug("Sending transcation!", newtx.Hash())
	err = ts.ethclient.SendTransaction(ctx, sigTx)

	if err != nil {
		return nil, err
	}

	return newtx, nil
}

func (ts *TxUtilService) checkTransactionReceipt(ctx context.Context, tx *types.Transaction) (*types.Receipt, error) {
	timer := time.NewTimer(1 * time.Second)

	for {
		select {
		case <-timer.C:
			start := time.Now()
			receipt, err := ts.ethclient.TransactionReceipt(ctx, tx.Hash())
			if err != nil {
				if err == ethereum.NotFound {
					continue
				} else {
					return receipt, err
				}
			}
			//TODO(raneet10): This should possibly configurable
			if time.Since(start) >= 10*time.Second {
				return receipt, ErrorTimeOut
			}

			return receipt, err

		case <-ctx.Done():
			break
		}

	}
}

func (ts *TxUtilService) GetGasfee() error {
	resp, err := http.Get("https://gasstation-mainnet.matic.network/")
	if err != nil {
		return err
	}
	gasinfo, err := io.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	err = json.Unmarshal(gasinfo, &ts.GasInfo)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}
