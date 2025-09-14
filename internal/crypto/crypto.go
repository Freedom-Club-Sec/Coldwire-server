package crypto

import (
    "fmt"

    "github.com/golang-jwt/jwt/v5"
    "github.com/cloudflare/circl/sign/mldsa/mldsa87"
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


func PublicKeyFromBytes(publicKeyBytes []byte) (*mldsa87.PublicKey, error) {
    publicKey := new(mldsa87.PublicKey)
    if err := publicKey.UnmarshalBinary(publicKeyBytes); err != nil {
        return nil, fmt.Errorf("Failed to unmarshal public key: %w", err)
    }
    return publicKey, nil
}

func VerifySignature(publicKey *mldsa87.PublicKey, data []byte, ctx []byte, signature []byte) bool {
    return mldsa87.Verify(publicKey, data, ctx, signature)
}
