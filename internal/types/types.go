package types

type AuthenticateInitRequest struct {
    PublicKey string `json:"public_key"`
    UserID    string `json:"user_id"`
}

type AuthenticateInitResponse struct {
    Challenge string `json:"challenge"`
}

type AuthenticateVerificationRequest struct {
    Challenge string `json:"challenge"`
    Signature string `json:"signature"`
}

type AuthenticateVerificationResponse struct {
    UserID string `json:"user_id"`
    Token  string `json:"token"`
}

