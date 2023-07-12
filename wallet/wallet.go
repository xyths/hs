package wallet

import (
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"
	"io/ioutil"
	"os"
)

type Wallet struct {
	w           *hdwallet.Wallet
	accounts    []accounts.Account
	privateKeys []*ecdsa.PrivateKey
	publicKeys  []*ecdsa.PublicKey
	addresses   []common.Address
}

func loadMnemonic(filename string) (string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func NewWallet(mnemonicFile string) (*Wallet, error) {
	mnemonic, err := loadMnemonic(mnemonicFile)
	if err != nil {
		return nil, err
	}

	wallet, err := hdwallet.NewFromMnemonic(mnemonic)
	if err != nil {
		return nil, err
	}

	return &Wallet{w: wallet}, nil
}

func (w *Wallet) Account(index int) (accounts.Account, error) {
	if index >= len(w.accounts) {
		if err := w.derive(index); err != nil {
			return accounts.Account{}, err
		}
	}
	return w.accounts[index], nil
}

func (w *Wallet) PrivateKey(index int) (*ecdsa.PrivateKey, error) {
	if index >= len(w.accounts) {
		if err := w.derive(index); err != nil {
			return nil, err
		}
	}
	return w.privateKeys[index], nil
}

func (w *Wallet) PublicKey(index int) (*ecdsa.PublicKey, error) {
	if index >= len(w.accounts) {
		if err := w.derive(index); err != nil {
			return nil, err
		}
	}
	return w.publicKeys[index], nil
}

func (w *Wallet) Address(index int) (common.Address, error) {
	if index >= len(w.accounts) {
		if err := w.derive(index); err != nil {
			return common.Address{}, err
		}
	}
	return w.addresses[index], nil
}

func (w *Wallet) derive(index int) error {
	s := len(w.accounts)
	for i := s; i <= index; i++ {
		path := hdwallet.MustParseDerivationPath(fmt.Sprintf("m/44'/60'/0'/0/%d", i))
		account, err := w.w.Derive(path, false)
		if err != nil {
			return err
		}
		privateKey, err := w.w.PrivateKey(account)
		if err != nil {
			return err
		}
		publicKey, err := w.w.PublicKey(account)
		if err != nil {
			return err
		}
		address := crypto.PubkeyToAddress(*publicKey)
		w.accounts = append(w.accounts, account)
		w.privateKeys = append(w.privateKeys, privateKey)
		w.publicKeys = append(w.publicKeys, publicKey)
		w.addresses = append(w.addresses, address)
	}
	return nil
}
