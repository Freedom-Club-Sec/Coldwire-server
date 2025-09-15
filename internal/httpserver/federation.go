package httpserver

import (
    "encoding/json"
    "log/slog"
    "net/http"
    "time"

    "github.com/cloudflare/circl/sign/mldsa/mldsa87"
    "github.com/Freedom-Club-Sec/Coldwire-server/internal/types"
    "github.com/Freedom-Club-Sec/Coldwire-server/internal/crypto"
)



func (s *Server) federationInfoHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    slog.Info("Received federation information fetch request")

    todayUTC := time.Now().UTC().Truncate(24 * time.Hour)

	refetchDate := todayUTC.AddDate(0, 0, 1).Format("2006-01-02")

	dataToSign := append([]byte(s.Cfg.DomainOrIP), []byte(refetchDate)...)

    ourPrivateKey, err := crypto.PrivateKeyFromBytes(s.Cfg.DSAPrivateKey)
    if err != nil {
        slog.Error("Error while parsing our private key.", "error", err)
        http.Error(w, "Error while processing request.", http.StatusBadRequest)
        return
    }

    ourPublicKey := ourPrivateKey.Public()

    ourPublicKeyEncoded, err := ourPublicKey.(*mldsa87.PublicKey).MarshalBinary()
    if err != nil {
        slog.Error("Error while encoding our public key.", "error", err)
        http.Error(w, "Error while processing request.", http.StatusBadRequest)
        return
    }



    signature, err := crypto.CreateSignature(ourPrivateKey, dataToSign, nil)
    if err != nil {
        slog.Error("Error while creating the signature.", "dataToSign", dataToSign, "error", err)
        http.Error(w, "Error while processing request.", http.StatusBadRequest)
        return
    }

    resp := types.FederationResponse{
        Signature: signature,
        PublicKey: ourPublicKeyEncoded,
        RefetchDate: refetchDate,
	}

    if err := json.NewEncoder(w).Encode(resp); err != nil {
        slog.Error("Error while encoding response.", "resp", resp, "error", err)
        http.Error(w, "Error while processing request.", http.StatusBadRequest)
	}
}
