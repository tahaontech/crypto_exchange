package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

func transferETH(client *ethclient.Client, from *ecdsa.PrivateKey, to common.Address, amount *big.Int) error {
	ctx := context.Background()

	publicKey := from.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return fmt.Errorf("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return err
	}

	gasLimit := uint64(21000) // in units
	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		return err
	}

	tx := types.NewTransaction(nonce, to, amount, gasLimit, gasPrice, nil)
	chainID := big.NewInt(1337)
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), from)
	if err != nil {
		return err
	}

	return client.SendTransaction(ctx, signedTx)
}

func GetBalance(client *ethclient.Client, address common.Address) (*big.Int, error) {
	ctx := context.Background()
	return client.BalanceAt(ctx, address, nil)
}
