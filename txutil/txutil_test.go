package txutil

import (
	"context"
	"math/big"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/params"
	log "github.com/inconshreveable/log15"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
)

func TestTransactionFlow(t *testing.T) {
	client, err := ethclient.Dial("https://polygon-rpc.com/")
	ctx := context.Background()
	nonce, err := client.NonceAt(ctx, common.HexToAddress("0xf89154D7A42c5E9f77884511586A9db4618683C5"), nil)

	addresses := []common.Address{common.HexToAddress("0x94f510C328C85CA726E3d79cefC9E798aF059b1E"),
		common.HexToAddress("0x823c5228B8aFCA18c54d291483E1B0d87fc32b0E"),
		common.HexToAddress("0x0E2974345646146144c96063c62D0613Ed97c432"),
		common.HexToAddress("0x7C590568dBAf309a1fB0c3C6aE1f7a32c0F2Ed1B"),
		common.HexToAddress("0x2261fDA0234D405F843f9957621aA74d74a6d4Ad"),
		common.HexToAddress("0x7d2732Ff636DCBA2792bEdC9c0C1aC9c1824D1Fe"),
		common.HexToAddress("0xc7f10250be7D9008368D885caeE175e98c88A65f"),
		common.HexToAddress("0x8F67d6BEc54010C6f598A35785d7ABaEFf8dc7Dd"),
		common.HexToAddress("0x85CA1fE9AD326be2bc1644c0af4a1C520Ee5d7b6"),
		common.HexToAddress("0x9d4C08656958F7011757ad14d370bf2183FeF937")}

	err = godotenv.Load("../.env")
	require.NoError(t, err)

	pk := os.Getenv("PRIVATE_KEY")
	privatekey := crypto.ToECDSAUnsafe(common.FromHex(pk))

	tc := NewTxUtilService(privatekey, client, ctx, log.New("module", "TxUtilService"))

	for i, address := range addresses {
		tx := types.NewTransaction(nonce+uint64(i),
			address,
			big.NewInt(20000000*params.GWei),
			uint64(21000),
			big.NewInt(20*params.GWei),
			nil)

		err = tc.HandleTransactions(ctx, tx)
		require.NoError(t, err)
	}

	require.NoError(t, err)

	/*for tx, status := range tc.Txreceipt {
		tc.logger.Debug("tx", tx, "status", status)
	}*/

}
