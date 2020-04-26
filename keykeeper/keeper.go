package keykeeper

import (
	"bytes"
	"errors"
	"encoding/json"
	"fmt"
	"time"
	"sync"
	"io"
	"io/ioutil"
	"os"
	"crypto/sha256"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"

	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	bip39 "github.com/cosmos/go-bip39"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/crypto/keys/hd"
)

const (
	AesNonceLength = 12
)

type AccountInfo struct {
	Memo              string `json:"memo"`
	Address           string `json:"address"`
	PassphraseCksum   []byte `json:"passphrase_cksum"`
	EncryptedMnemonic []byte `json:"encrypted_mnemonic"`
}

func NewAccountInfo(memo, mnemonic, passphrase string) AccountInfo {
	sum1 := sha256.Sum256([]byte(passphrase))
	sum2 := sha256.Sum256(sum1[:])
	encMnemonic, nonce := AesGcmEncrypt(sum1[:], mnemonic)
	_, _, addr := getAllFromMnemonic(mnemonic)
	return AccountInfo{
		Memo:              memo,
		Address:           addr,
		PassphraseCksum:   sum2[:],
		EncryptedMnemonic: append(nonce, []byte(encMnemonic)...),
	}
}

func (acc AccountInfo) CheckPassphrase(passphrase string) error {
	sum1 := sha256.Sum256([]byte(passphrase))
	sum2 := sha256.Sum256(sum1[:])
	if !bytes.Equal(sum2[:], acc.EncryptedMnemonic) {
		return errors.New("Passphrase's checksum does not match")
	}
	return nil
}

func getAllFromMnemonic(mnemonic string) (privk secp256k1.PrivKeySecp256k1, pubk secp256k1.PubKeySecp256k1, addr string) {
	seed := bip39.NewSeed(mnemonic, DefaultBIP39Passphrase)
	fullHdPath := hd.NewFundraiserParams(0, DefaultCoinType, 0) //account=0 addressIdx=0
	masterPriv, ch := hd.ComputeMastersFromSeed(seed)
	derivedPriv, err := hd.DerivePrivateKeyForPath(masterPriv, ch, fullHdPath.String())
	if err != nil {
		panic(err)
	}
	privk = secp256k1.PrivKeySecp256k1(derivedPriv)
	pubk = privk.PubKey().(secp256k1.PubKeySecp256k1)
	addrBytes := pubk.Address()
	addr = sdk.AccAddress(addrBytes).String()
	return
}

type MyKeyBase struct {
	mtx              sync.RWMutex
	openedFile       *os.File
	cachedPassphrase map[string]string
	Accounts         []AccountInfo
}

func (kb *MyKeyBase) GetCachedPassphrase(addr string) (res string, ok bool) {
	kb.mtx.RLock()
	defer kb.mtx.RUnlock()
	res, ok = kb.cachedPassphrase[addr]
	return
}

func (kb *MyKeyBase) AddCachedPassphrase(addr, passphrase string) error {
	kb.mtx.RLock()
	defer kb.mtx.RUnlock()
	accInfo, ok := kb.GetAccountInfo(addr)
	if !ok {
		return errors.New("No such account")
	}
	err := accInfo.CheckPassphrase(passphrase)
	if err != nil {
		return err
	}
	oldPass, ok := kb.cachedPassphrase[addr]
	if ok && oldPass  == passphrase {
		return nil
	}
	kb.cachedPassphrase[addr] = passphrase

	// delete the cached passphrase after 5 minutes
	timer := time.NewTimer(time.Second*5*60)
	go func() {
		<-timer.C
		kb.mtx.Lock()
		defer kb.mtx.Unlock()
		delete(kb.cachedPassphrase, addr)
	}()
	return nil
}

func (kb *MyKeyBase) AddAccount(accInfo AccountInfo) {
	kb.mtx.Lock()
	defer kb.mtx.Unlock()
	for i := range kb.Accounts {
		if accInfo.Address == kb.Accounts[i].Address {
			kb.Accounts[i] = accInfo
			return
		}
	}
	kb.Accounts = append(kb.Accounts, accInfo)
}

func (kb *MyKeyBase) DeleteAccount(addr string) {
	kb.mtx.Lock()
	defer kb.mtx.Unlock()
	if len(kb.Accounts) == 0 {
		return
	}
	idx := -1
	for i, acc := range kb.Accounts {
		if acc.Address == addr {
			idx = i
			break
		}
	}
	if idx == -1 {
		return
	}
	if idx != len(kb.Accounts)-1 {
		kb.Accounts[idx] = kb.Accounts[len(kb.Accounts)-1]
	}
	kb.Accounts = kb.Accounts[:len(kb.Accounts)-1]
}

func (kb *MyKeyBase) GetAccountInfo(addr string) (AccountInfo, bool) {
	kb.mtx.RLock()
	defer kb.mtx.RUnlock()
	for _, acc := range kb.Accounts {
		if acc.Address == addr {
			return acc, true
		}
	}
	return AccountInfo{}, false
}

func (kb *MyKeyBase) Init(openedFile *os.File, accounts []AccountInfo) {
	kb.mtx.Lock()
	defer kb.mtx.Unlock()
	kb.openedFile = openedFile
	kb.Accounts = accounts
	kb.cachedPassphrase = make(map[string]string)
}

func (kb *MyKeyBase) ChangePassphrase(addr, oldPassphrase, newPassphrase string) error {
	accInfo, ok := kb.GetAccountInfo(addr)
	if !ok {
		return errors.New("No such address")
	}
	err := accInfo.CheckPassphrase(oldPassphrase)
	if err != nil {
		return errors.New("Old passphrase is incorrect")
	}
	ciphertext := accInfo.EncryptedMnemonic[AesNonceLength:]
	nonce := accInfo.EncryptedMnemonic[:AesNonceLength]
	sum1 := sha256.Sum256([]byte(oldPassphrase))
	mnemonic := AesGcmDecrypt(sum1[:], ciphertext, nonce)
	accInfo = NewAccountInfo(accInfo.Memo, mnemonic, newPassphrase)
	kb.AddAccount(accInfo)
	return nil
}

func (kb *MyKeyBase) GetMnemonic(addr, passphrase string) (string, error) {
	accInfo, ok := kb.GetAccountInfo(addr)
	if !ok {
		return "", errors.New("No such account")
	}
	ciphertext := accInfo.EncryptedMnemonic[AesNonceLength:]
	nonce := accInfo.EncryptedMnemonic[:AesNonceLength]
	sum1 := sha256.Sum256([]byte(passphrase))
	mnemonic := AesGcmDecrypt(sum1[:], ciphertext, nonce)
	return mnemonic, nil
}

func (kb *MyKeyBase) Sign(addr, passphrase string, msg []byte) (sig []byte, pubk secp256k1.PubKeySecp256k1, err error) {
	mnemonic, err := kb.GetMnemonic(addr, passphrase)
	privk, pubk, _ := getAllFromMnemonic(mnemonic)
	sig, err = privk.Sign(msg)
	return
}

func (kb *MyKeyBase) Save() error {
	err := kb.openedFile.Truncate(0)
	if err != nil {
		return err
	}
	b, err := json.Marshal(kb.Accounts)
	if err != nil {
		return err
	}
	_, err = kb.openedFile.Write(b)
	if err != nil {
		return err
	}
	return kb.openedFile.Sync()
}

func (kb *MyKeyBase) IsOpen() bool {
	kb.mtx.RLock()
	defer kb.mtx.RUnlock()
	return kb.openedFile != nil
}

func (kb *MyKeyBase) GetStringItems() (items []string) {
	kb.mtx.RLock()
	defer kb.mtx.RUnlock()
	for _, accInfo := range kb.Accounts {
		items = append(items, accInfo.Address+": "+accInfo.Memo)
	}
	return
}

func (kb *MyKeyBase) Close() {
	kb.mtx.Lock()
	defer kb.mtx.Unlock()
	if kb.openedFile != nil {
		kb.openedFile.Close()
		kb.openedFile = nil
	}
}

// ================================================
var KB MyKeyBase

func CloseKeybase() {
	KB.Close()
}

func KeybaseOpened() bool {
	return KB.IsOpen()
}

func GetMnemonic(addr, passphrase string) (string, error) {
	return KB.GetMnemonic(addr, passphrase)
}

func GetCachedPassphrase(addr string) (res string, ok bool) {
	return KB.GetCachedPassphrase(addr)
}

func AddCachedPassphrase(addr, passphrase string) error {
	return KB.AddCachedPassphrase(addr, passphrase)
}

func HasAccount(addr string) bool {
	_, ok := KB.GetAccountInfo(addr)
	return ok
}

func OpenKeybase(fname string) error {
	KB.Close()
	info, err := os.Stat(fname)
	fileNotExists := os.IsNotExist(err)
	fmt.Printf("1 %#v\n", err)
	if err != nil && !fileNotExists {
		return err
	}
	if info != nil && info.IsDir() {
		return errors.New(fname+" is not a plain file")
	}
	if fileNotExists { //So a new empty file is created
		f, err := os.Create(fname)
		if err != nil {
			return err
		}
		f.Close()
	}
	openedFile, err := os.OpenFile(fname, os.O_RDWR, 0644)
	fmt.Printf("2 %#v\n", err)
	if err != nil {
		return err
	}
	if fileNotExists {
		KB.Init(openedFile, []AccountInfo{})
		return nil
	}
	content, err := ioutil.ReadAll(openedFile)
	if err != nil {
		KB.openedFile = nil
		openedFile.Close()
		return err
	}
	var accounts []AccountInfo
	err = json.Unmarshal(content, &accounts)
	if err != nil {
		KB.openedFile = nil
		openedFile.Close()
		return err
	}
	KB.Init(openedFile, accounts)
	return nil
}

func CreateAccount(memo, mnemonic, passphrase string) (AccountInfo, error) {
	accInfo := NewAccountInfo(memo, mnemonic, passphrase)
	KB.AddAccount(accInfo)
	err := KB.Save()
	return accInfo, err
}

func ChangePassphrase(addr, oldPassphrase, newPassphrase string) error {
	err := KB.ChangePassphrase(addr, oldPassphrase, newPassphrase)
	if err != nil {
		return err
	}
	return KB.Save()
}

func DeleteAccount(addr, passphrase string) error {
	accInfo, ok := KB.GetAccountInfo(addr)
	if !ok {
		return errors.New("No such account")
	}
	err := accInfo.CheckPassphrase(passphrase)
	if err != nil {
		return err
	}
	KB.DeleteAccount(addr)
	return KB.Save()
}

func Sign(name, passphrase string, msg []byte) (string, error) {
	sig, pub, err := KB.Sign(name, passphrase, msg)
	if err != nil {
		return "", err
	}
	stdSign := auth.StdSignature{pub, sig}
	out, err := gCdc.MarshalJSON(stdSign)
	if err != nil {
		return "", err
	}
	return string(out), nil
}


// ================================================
var gCdc = codec.New()

func initCodec() {
	gCdc.RegisterInterface((*crypto.PubKey)(nil), nil)
	gCdc.RegisterInterface((*crypto.PrivKey)(nil), nil)
	gCdc.RegisterInterface((*sdk.Msg)(nil), nil)
	gCdc.RegisterConcrete(secp256k1.PubKeySecp256k1{}, "tendermint/PubKeySecp256k1", nil)
	gCdc.RegisterConcrete(secp256k1.PrivKeySecp256k1{}, "tendermint/PrivKeySecp256k1", nil)
}

func init() {
	initCodec()
	bench32MainPrefix := "coinex"
	bench32PrefixAccAddr := bench32MainPrefix
	// bench32PrefixAccPub defines the bench32 prefix of an account's public key
	bench32PrefixAccPub := bench32MainPrefix + sdk.PrefixPublic
	// bench32PrefixValAddr defines the bench32 prefix of a validator's operator address
	bench32PrefixValAddr := bench32MainPrefix + sdk.PrefixValidator + sdk.PrefixOperator
	// bench32PrefixValPub defines the bench32 prefix of a validator's operator public key
	bench32PrefixValPub := bench32MainPrefix + sdk.PrefixValidator + sdk.PrefixOperator + sdk.PrefixPublic
	// bench32PrefixConsAddr defines the bench32 prefix of a consensus node address
	bench32PrefixConsAddr := bench32MainPrefix + sdk.PrefixValidator + sdk.PrefixConsensus
	// bench32PrefixConsPub defines the bench32 prefix of a consensus node public key
	bench32PrefixConsPub := bench32MainPrefix + sdk.PrefixValidator + sdk.PrefixConsensus + sdk.PrefixPublic

	config := sdk.GetConfig()
	config.SetCoinType(DefaultCoinType)
	config.SetBech32PrefixForAccount(bench32PrefixAccAddr, bench32PrefixAccPub)
	config.SetBech32PrefixForValidator(bench32PrefixValAddr, bench32PrefixValPub)
	config.SetBech32PrefixForConsensusNode(bench32PrefixConsAddr, bench32PrefixConsPub)
	config.Seal()
}

// ================================================
// AesGcmEncrypt takes an encryption key and a plaintext string and encrypts it with AES256 in GCM mode, 
// which provides authenticated encryption. Returns the ciphertext and the used nonce.
// len(key) must be 32, to select AES256
func AesGcmEncrypt(key []byte, plaintext string) (ciphertext, nonce []byte) {
	plaintextBytes := []byte(plaintext)

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err.Error())
	}

	// Never use more than 2^32 random nonces with a given key because of the risk of a repeat.
	nonce = make([]byte, AesNonceLength)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		panic(err.Error())
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}

	ciphertext = aesgcm.Seal(nil, nonce, plaintextBytes, nil)
	//fmt.Printf("Ciphertext: %x\n", ciphertext)
	//fmt.Printf("Nonce: %x\n", nonce)

	return
}

// AesGcmDecrypt takes an decryption key, a ciphertext and the corresponding nonce, 
// and decrypts it with AES256 in GCM mode. Returns the plaintext string.
// len(key) must be 32, to select AES256
func AesGcmDecrypt(key, ciphertext, nonce []byte) (plaintext string) {
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err.Error())
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}

	plaintextBytes, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		panic(err.Error())
	}

	plaintext = string(plaintextBytes)
	//fmt.Printf("%s\n", plaintext)

	return
}
