package crypto

import (
	"fmt"

	"github.com/cloudflare/circl/sign/mldsa/mldsa87"
	"github.com/golang-jwt/jwt/v5"
)

func CreateJWTToken(claims map[string]interface{}, JWTSecret []byte) (string, error) {
	tokenClaims := jwt.MapClaims{}
	for k, v := range claims {
		tokenClaims[k] = v
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, tokenClaims)
	tokenString, err := token.SignedString(JWTSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func VerifyJWT(tokenString string, JWTSecret []byte) (*jwt.Token, jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return JWTSecret, nil
	})

	if err != nil {
		return nil, nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, nil, fmt.Errorf("invalid claims type")
	}

	return token, claims, nil
}

func CreateDSAKeyPair() (*mldsa87.PublicKey, *mldsa87.PrivateKey, error) {
	return mldsa87.GenerateKey(nil)
}

func PrivateKeyFromBytes(privateKeyBytes []byte) (*mldsa87.PrivateKey, error) {
	privateKey := new(mldsa87.PrivateKey)
	if err := privateKey.UnmarshalBinary(privateKeyBytes); err != nil {
		return nil, fmt.Errorf("Failed to unmarshal private key: %w", err)
	}
	return privateKey, nil
}

func PublicKeyFromBytes(publicKeyBytes []byte) (*mldsa87.PublicKey, error) {
	publicKey := new(mldsa87.PublicKey)
	if err := publicKey.UnmarshalBinary(publicKeyBytes); err != nil {
		return nil, fmt.Errorf("Failed to unmarshal public key: %w", err)
	}
	return publicKey, nil
}

func CreateSignature(privateKey *mldsa87.PrivateKey, data []byte, ctx []byte) ([]byte, error) {
	buf := make([]byte, mldsa87.SignatureSize)
	err := mldsa87.SignTo(privateKey, data, ctx, false, buf)
	if err != nil {
		return nil, err
	}

	return buf, err
}

func VerifySignature(publicKey *mldsa87.PublicKey, data []byte, ctx []byte, signature []byte) bool {
	return mldsa87.Verify(publicKey, data, ctx, signature)
}
