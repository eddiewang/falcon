package libwallet

import (
	"bytes"
	"encoding/hex"

	"github.com/btcsuite/btcd/wire"
	"github.com/pkg/errors"
)

type MuunAddress interface {
	Version() int
	DerivationPath() string
	Address() string
}

type Outpoint interface {
	TxId() []byte
	Index() int
	Amount() int64
}

type InputSubmarineSwap interface {
	RefundAddress() string
	PaymentHash256() []byte
	ServerPublicKey() []byte
	LockTime() int64
}

type Input interface {
	OutPoint() Outpoint
	Address() MuunAddress
	UserSignature() []byte
	MuunSignature() []byte
	SubmarineSwap() InputSubmarineSwap
}

type PartiallySignedTransaction struct {
	tx     *wire.MsgTx
	inputs []Input
}

type Transaction struct {
	Hash  string
	Bytes []byte
}

func NewPartiallySignedTransaction(hexTx string) (*PartiallySignedTransaction, error) {

	rawTx, err := hex.DecodeString(hexTx)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to decode hex tx")
	}

	tx := wire.NewMsgTx(0)
	err = tx.BtcDecode(bytes.NewBuffer(rawTx), 0, wire.WitnessEncoding)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to decode tx")
	}

	return &PartiallySignedTransaction{tx: tx, inputs: []Input{}}, nil
}

func (p *PartiallySignedTransaction) AddInput(input Input) {
	p.inputs = append(p.inputs, input)
}

func (p *PartiallySignedTransaction) Sign(key *HDPrivateKey, muunKey *HDPublicKey) (*Transaction, error) {

	for i, input := range p.inputs {

		derivedKey, err := key.DeriveTo(input.Address().DerivationPath())
		if err != nil {
			return nil, errors.Wrapf(err, "failed to derive user key")
		}

		derivedMuunKey, err := muunKey.DeriveTo(input.Address().DerivationPath())
		if err != nil {
			return nil, errors.Wrapf(err, "failed to derive muun key")
		}

		var txIn *wire.TxIn

		switch AddressVersion(input.Address().Version()) {
		case addressV1:
			txIn, err = signInputV1(input, i, p.tx, derivedKey)
		case addressV2:
			txIn, err = signInputV2(input, i, p.tx, derivedKey, derivedMuunKey)
		case addressV3:
			txIn, err = signInputV3(input, i, p.tx, derivedKey, derivedMuunKey)
		case addressSubmarineSwap:
			txIn, err = signInputSubmarineSwap(input, i, p.tx, derivedKey, derivedMuunKey)
		default:
			return nil, errors.Errorf("cant sign transaction of version %v", input.Address().Version())
		}

		if err != nil {
			return nil, errors.Wrapf(err, "failed to sign input using version %v", input.Address().Version())
		}

		p.tx.TxIn[i] = txIn
	}

	var writer bytes.Buffer
	err := p.tx.BtcEncode(&writer, 0, wire.WitnessEncoding)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to encode tx")
	}

	return &Transaction{
		Hash:  p.tx.TxHash().String(),
		Bytes: writer.Bytes(),
	}, nil

}
