package httpserver

import (
    "encoding/json"
    "log/slog"
    "net/http"
    "time"
    "io"

    "github.com/cloudflare/circl/sign/mldsa/mldsa87"
    "github.com/Freedom-Club-Sec/Coldwire-server/internal/types"
    "github.com/Freedom-Club-Sec/Coldwire-server/internal/crypto"
    "github.com/Freedom-Club-Sec/Coldwire-server/internal/utils"
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

    resp := types.FederationInfoResponse{
        Signature: signature,
        PublicKey: ourPublicKeyEncoded,
        RefetchDate: refetchDate,
	}

    if err := json.NewEncoder(w).Encode(resp); err != nil {
        slog.Error("Error while encoding response.", "resp", resp, "error", err)
        http.Error(w, "Error while processing request.", http.StatusBadRequest)
	}
}


func (s *Server) federationSendHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    slog.Info("Received federation send request")


    err := r.ParseMultipartForm(3 << 20) // 3 MB max memory
    if err != nil {
        slog.Error("Error while parsing request form.", "error", err)
        http.Error(w, "Failed to parse form.", http.StatusBadRequest)
        return
    }

    metadataStr := r.FormValue("metadata")
    if metadataStr == "" {
        slog.Error("Missing metadata from request.")
        http.Error(w, "Missing metadata", http.StatusBadRequest)
        return
    }

    
    var metadata types.FederationSendRequest 

    if err := json.Unmarshal([]byte(metadataStr), &metadata); err != nil {
        slog.Error("Error while parsing request JSON metadata.", "error", err)
        http.Error(w, "Invalid JSON metadata.", http.StatusBadRequest)
        return
    }

    if metadata.Recipient == "" {
        slog.Error("Empty recipient from request metadata.")
        http.Error(w, "Missing recipient in metadata", http.StatusBadRequest)
        return
    }


    if metadata.Sender == "" {
        http.Error(w, "Missing sender in metadata", http.StatusBadRequest)
        return
    }

    if metadata.Url == "" {
        http.Error(w, "Missing url in metadata", http.StatusBadRequest)
        return
    }

    if !utils.IsAllDigits(metadata.Sender)  {
        http.Error(w, "Malformed sender.", http.StatusBadRequest)
        return
    }

    if !utils.IsAllDigits(metadata.Recipient)  {
        http.Error(w, "Malformed recipient.", http.StatusBadRequest)
        return
    }

    if !utils.IsValidDomainOrIP(metadata.Url, s.Cfg.BlacklistedIPs, s.Cfg.BlacklistedDomains)  {
        http.Error(w, "Invalid url.", http.StatusBadRequest)
        return
    }

    file, _, err := r.FormFile("blob")
    if err != nil {
        http.Error(w, "Failed to read file: "+err.Error(), http.StatusBadRequest)
        return
    }
    defer file.Close()

    blobData, err := io.ReadAll(file)
    if err != nil {
        http.Error(w, "Failed to read blob data: "+err.Error(), http.StatusBadRequest)
        return
    }

    if len(blobData) == 0 {
        http.Error(w, "Empty blob is not allowed", http.StatusBadRequest)
        return
    }

    if err := s.DbSvcs.DataService.FederationProcessor(metadata.Sender, metadata.Recipient, metadata.Url, blobData); err != nil {
        slog.Error("Failure when attempted to process federation request.", "sender", metadata.Sender, "recipient", "url", metadata.Url, metadata.Recipient, "error", err)
        http.Error(w, "Failed to process data.", http.StatusBadRequest)
        return
    }


}
