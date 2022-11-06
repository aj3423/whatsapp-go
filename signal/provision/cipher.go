package provision

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"wa/signal/cipher"
	"wa/signal/ecc"
	"wa/signal/kdf"
	"wa/signal/keys/root"
	"wa/signal/util/bytehelper"
)

type ProvisionMessage struct {
	IdentityKeyPublic  []byte `json:"identity_key_public"`
	IdentityKeyPrivate []byte `json:"identity_key_private"`
	UserId             string `json:"user_id"`
	ProvisioningCode   string `json:"provisioning_code"`
	ProfileKey         []byte `json:"profile_key"`
}

type ProvisionEnvelope struct {
	PublicKey []byte `json:"public_key"`
	Body      []byte `json:"body"`
}

func verifyMAC(key, input, mac []byte) bool {
	m := hmac.New(sha256.New, key)
	m.Write(input)
	return hmac.Equal(m.Sum(nil), mac)
}

func Decrypt(privateKey string, content string) (string, error) {
	ourPrivateKey, err := base64.StdEncoding.DecodeString(privateKey)
	if err != nil {
		return "", err
	}
	envelopeDecode, err := base64.StdEncoding.DecodeString(content)
	if err != nil {
		return "", err
	}

	var envelope ProvisionEnvelope
	if err := json.Unmarshal(envelopeDecode, &envelope); err != nil {
		return "", err
	}

	publicKeyable, _ := ecc.DecodePoint(envelope.PublicKey, 0)
	masterEphemeral := publicKeyable.PublicKey()
	message := envelope.Body
	if message[0] != 1 {
		return "", fmt.Errorf("Bad version number on ProvisioningMessage %s", err.Error())
	}

	iv := message[1 : 16+1]
	mac := message[len(message)-32:]
	ivAndCiphertext := message[0 : len(message)-32]
	cipherText := message[16+1 : len(message)-32]

	sharedSecret := kdf.CalculateSharedSecret(masterEphemeral, bytehelper.SliceToArray(ourPrivateKey))
	derivedSecretBytes, err := kdf.DeriveSecrets(sharedSecret[:], nil, []byte("Mixin Provisioning Message"), root.DerivedSecretsSize)
	if err != nil {
		return "", err
	}
	aesKey := derivedSecretBytes[:32]
	macKey := derivedSecretBytes[32:]

	if !verifyMAC(macKey, ivAndCiphertext, mac) {
		return "", fmt.Errorf("Verify Mac failed")
	}
	plaintext, err := cipher.DecryptCbc(iv, aesKey, cipherText)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}
