// Copyright 2010-2025 the original author or authors.

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

// 	http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package autoconfig

import (
	"bytes"
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"reflect"

	"github.com/stella-go/siu/config"
	"github.com/stella-go/siu/interfaces"
)

const (
	CipherKey           = "cipher"
	CipherDisableKey    = CipherKey + ".disable"
	CipherKeyKey        = CipherKey + ".key"
	CipherHmacKeyKey    = CipherKey + ".hmac-key"
	CipherPublicKeyKey  = CipherKey + ".public-key"
	CipherPrivateKeyKey = CipherKey + ".private-key"
	CipherOrder         = 100
)

type AutoCipher struct {
	Conf   config.TypedConfig `@siu:"name='environment',default='type'"`
	cipher interfaces.Cipher
}

func (p *AutoCipher) Condition() bool {
	_, ok1 := p.Conf.Get(CipherKey)
	v, ok2 := p.Conf.GetBool(CipherDisableKey)

	if ok2 && v {
		return false
	}
	return ok1
}

func (p *AutoCipher) OnStart() error {
	skey := p.Conf.GetStringOr(CipherKeyKey, "")
	shmacKey := p.Conf.GetStringOr(CipherHmacKeyKey, "")
	spublicKey := p.Conf.GetStringOr(CipherPublicKeyKey, "")
	sprivateKey := p.Conf.GetStringOr(CipherPrivateKeyKey, "")
	cipher, err := NewCipherImpl(skey, shmacKey, spublicKey, sprivateKey)
	if err != nil {
		return err
	}
	p.cipher = cipher
	return nil
}

func (p *AutoCipher) OnStop() error {
	return nil
}

func (*AutoCipher) Order() int {
	return CipherOrder
}

func (*AutoCipher) Name() string {
	return CipherKey
}

func (p *AutoCipher) Named() map[string]interface{} {
	return map[string]interface{}{
		CipherKey: p.cipher,
	}
}

func (p *AutoCipher) Typed() map[reflect.Type]interface{} {
	refType := reflect.TypeOf((*interfaces.Cipher)(nil)).Elem()
	return map[reflect.Type]interface{}{
		refType: p.cipher,
	}
}

type CipherImpl struct {
	Key        []byte
	HmacKey    []byte
	PublicKey  string
	PrivateKey *rsa.PrivateKey
}

func NewCipherImpl(skey string, shmacKey string, spublicKey string, sprivateKey string) (*CipherImpl, error) {
	var key, hmacKey []byte
	var publicKey string
	var privateKey *rsa.PrivateKey
	var err error

	if skey != "" {
		key, err = hex.DecodeString(skey)
		if err != nil {
			return nil, err
		}
		key = append(pad(key), bytes.Repeat([]byte{16}, 16)...)[:32]
	}

	if shmacKey != "" {
		hmacKey, err = hex.DecodeString(shmacKey)
		if err != nil {
			return nil, err
		}
		hmacKey = append(pad(hmacKey), bytes.Repeat([]byte{16}, 16)...)[:32]
	}

	if spublicKey != "" {
		publicKey = spublicKey
	}

	if sprivateKey != "" {
		block, _ := pem.Decode([]byte(sprivateKey))
		if block == nil {
			return nil, fmt.Errorf("failed to parse PEM block containing the key")
		}
		privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
	}

	return &CipherImpl{Key: key, HmacKey: hmacKey, PublicKey: publicKey, PrivateKey: privateKey}, nil
}

func (p *CipherImpl) Encrypt(s string) string {
	src := []byte(s)
	iv := p.Random(16)
	src = pad(src)
	dst := make([]byte, len(src))
	block, err := aes.NewCipher(p.Key)
	if err != nil {
		panic(err)
	}
	encryptor := cipher.NewCBCEncrypter(block, iv)
	encryptor.CryptBlocks(dst, src)
	return base64.StdEncoding.EncodeToString(append(iv, dst...))
}

func (p *CipherImpl) Decrypt(s string) (string, error) {
	enc, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return "", err
	}
	iv := enc[:16]
	enc = enc[16:]
	dec := make([]byte, len(enc))
	block, err := aes.NewCipher(p.Key)
	if err != nil {
		return "", err
	}
	decrypter := cipher.NewCBCDecrypter(block, iv)
	decrypter.CryptBlocks(dec, enc)
	src, err := unpad(dec)
	if err != nil {
		return "", err
	}
	return string(src), nil
}

func (p *CipherImpl) GetPublickey() string {
	return p.PublicKey
}

func (p *CipherImpl) PublickeyEncrypt(s string) string {
	enc, err := rsa.EncryptPKCS1v15(rand.Reader, &p.PrivateKey.PublicKey, []byte(s))
	if err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(enc)
}

func (p *CipherImpl) PrivateKeyDecrypt(s string) (string, error) {
	enc, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return "", err
	}
	src, err := rsa.DecryptPKCS1v15(rand.Reader, p.PrivateKey, enc)
	if err != nil {
		return "", err
	}
	return string(src), nil
}

func (p *CipherImpl) Sign(msg string) string {
	hashed := sha512.Sum512([]byte(msg))
	signature, err := rsa.SignPKCS1v15(rand.Reader, p.PrivateKey, crypto.SHA512, hashed[:])
	if err != nil {
		panic(err)
	}
	bts := signature
	return base64.StdEncoding.EncodeToString(bts)
}

func (p *CipherImpl) Verify(msg string, sign string) bool {
	bts, err := base64.StdEncoding.DecodeString(sign)
	if err != nil {
		return false
	}
	hashed := sha512.Sum512([]byte(msg))
	err = rsa.VerifyPKCS1v15(&p.PrivateKey.PublicKey, crypto.SHA512, hashed[:], bts)
	return err == nil
}

func (p *CipherImpl) Hash(s string) string {
	h := sha512.Sum512([]byte(s))
	return base64.StdEncoding.EncodeToString(h[:])
}

func (p *CipherImpl) Hmac(s string) string {
	h := hmac.New(sha512.New, p.HmacKey)
	h.Write([]byte(s))
	hash := h.Sum(nil)
	return base64.StdEncoding.EncodeToString(hash)
}

func (p *CipherImpl) Random(n int) []byte {
	bts := make([]byte, n)
	rand.Read(bts)
	return bts
}

func (p *CipherImpl) GenRsaKeyPair(length int) (string, string) {
	privateKey, err := rsa.GenerateKey(rand.Reader, length)
	if err != nil {
		panic(err)
	}
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPem := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		panic(err)
	}
	publicKeyPem := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: publicKeyBytes,
	})
	return string(privateKeyPem), string(publicKeyPem)
}

func pad(data []byte) []byte {
	tail := 16 - (len(data) & 15)
	return append(data, bytes.Repeat([]byte{byte(tail)}, tail)...)
}
func unpad(data []byte) ([]byte, error) {
	tail := data[len(data)-1]
	for i := len(data) - 1; i > len(data)-int(tail); i-- {
		if data[i] != tail {
			return nil, fmt.Errorf("unexpected padding")
		}
	}
	return data[:len(data)-int(tail)], nil
}
