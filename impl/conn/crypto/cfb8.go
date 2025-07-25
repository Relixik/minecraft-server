package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
)

type cfb8 struct {
	b       cipher.Block
	sr      []byte
	srEnc   []byte
	srPos   int
	decrypt bool
}

func newEncrypt(block cipher.Block, iv []byte) (cipher.Stream, error) {
	if len(iv) != block.BlockSize() {
		return nil, fmt.Errorf("cfb8.newEncrypt: IV length (%d) must equal block size (%d)", len(iv), block.BlockSize())
	}

	return newCFB8(block, iv, false), nil
}

func newDecrypt(block cipher.Block, iv []byte) (cipher.Stream, error) {
	if len(iv) != block.BlockSize() {
		return nil, fmt.Errorf("cfb8.newDecrypt: IV length (%d) must equal block size (%d)", len(iv), block.BlockSize())
	}

	return newCFB8(block, iv, true), nil
}

func newCFB8(block cipher.Block, iv []byte, decrypt bool) cipher.Stream {
	blockSize := block.BlockSize()
	if len(iv) != blockSize {
		return nil
	}

	x := &cfb8{
		b:       block,
		sr:      make([]byte, blockSize*4),
		srEnc:   make([]byte, blockSize),
		srPos:   0,
		decrypt: decrypt,
	}

	copy(x.sr, iv)

	return x
}

func (x *cfb8) XORKeyStream(dst, src []byte) {
	blockSize := x.b.BlockSize()

	for i := 0; i < len(src); i++ {
		x.b.Encrypt(x.srEnc, x.sr[x.srPos:x.srPos+blockSize])

		var c byte
		if x.decrypt {
			c = src[i]
			dst[i] = c ^ x.srEnc[0]
		} else {
			c = src[i] ^ x.srEnc[0]
			dst[i] = c
		}

		x.sr[x.srPos+blockSize] = c
		x.srPos++

		if x.srPos+blockSize == len(x.sr) {
			copy(x.sr, x.sr[x.srPos:])
			x.srPos = 0
		}
	}
}

func NewEncryptAndDecrypt(secret []byte) (encrypt cipher.Stream, decrypt cipher.Stream, err error) {
	block, err := aes.NewCipher(secret)

	if err != nil {
		return nil, nil, err
	}

	encrypt, err = newEncrypt(block, secret)
	if err != nil {
		return nil, nil, err
	}

	decrypt, err = newDecrypt(block, secret)
	if err != nil {
		return nil, nil, err
	}

	return
}
