package httpserver

import (
    "encoding/json"
    "log/slog"
    "net/http"

    "github.com/Freedom-Club-Sec/Coldwire-server/internal/crypto"
    "github.com/Freedom-Club-Sec/Coldwire-server/internal/utils"
    "github.com/Freedom-Club-Sec/Coldwire-server/internal/types"
)




func (s *Server) authenticateInitHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    var payload types.AuthenticateInitRequest 

    if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }

    slog.Info("Received authentication initializing request")

    if (payload.PublicKey == "" && payload.UserID == "") || (payload.PublicKey != "" && payload.UserID != "") {
        slog.Error("Request requires one field", "payload", payload)
        http.Error(w, "Request requires one field", http.StatusBadRequest)
        return
    } 

    if payload.UserID != "" {
        if len(payload.UserID) != 16 || !utils.IsAllDigits(payload.UserID) {
            slog.Error("Invalid UserID", "payload", payload)
            http.Error(w, "Invalid UserID", http.StatusBadRequest)
            return
        }
    }
    
    challengeEncoded, err := s.DbSvcs.UserService.AuthenticateInitProcessor(&payload)
    if err != nil {
        slog.Error("Error while processing request.",  "error", err, "payload", payload)
        http.Error(w, "Error while processing request.", http.StatusBadRequest)
        return
    }

    slog.Info("Created a new challenge", "challenge", challengeEncoded)
    
    resp := types.AuthenticateInitResponse{
		Challenge: challengeEncoded,
	}

    if err := json.NewEncoder(w).Encode(resp); err != nil {
        slog.Error("Error while encoding response.", "resp", resp, "error", err)
        http.Error(w, "Error while processing request.", http.StatusBadRequest)
	}

}

func (s *Server) authenticateVerificationHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    var payload types.AuthenticateVerificationRequest 

    if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }

    slog.Info("Received authentication verification request")

    if (payload.Signature == "" || payload.Challenge == "") {
        slog.Error("Request missing signature or challenge fields", "payload", payload)
        http.Error(w,  "Request missing signature or challenge fields", http.StatusBadRequest)
        return
    } 

    userId, publicKey, validSignature, err := s.DbSvcs.UserService.AuthenticateVerificationProcessor(&payload)
    if err != nil {
        slog.Error("Error while processing request.", "error", err, "payload", payload)
        http.Error(w, "Error while processing request.", http.StatusBadRequest)
        return
    }

    if validSignature {
        slog.Info("Challenge verification passed.", "challenge", payload.Challenge)
    } else {
        slog.Warn("Challenge verification failed!", "challenge", payload.Challenge)
    }

    if userId == "" {
        userId, err = s.DbSvcs.UserService.RegisterNewUser(publicKey)
        if err != nil {
            slog.Error("Failed to register new account, likely because of duplicated public-key.", "error", err)
            http.Error(w, "Error while processing request.", http.StatusBadRequest)
            return
        }
    }

    token, err := crypto.CreateJWTToken(map[string]interface{}{
        "user_id": userId,
    }, s.Cfg.JWTSecret)

    if err != nil {
        slog.Error("F.", "error", err)
        http.Error(w, "Error while processing request.", http.StatusBadRequest)
        return
    }
    


    resp := types.AuthenticateVerificationResponse{
		UserID: userId,
        Token: token,
	}

    if err := json.NewEncoder(w).Encode(resp); err != nil {
        slog.Error("Error while encoding response.", "resp", resp, "error", err)
        http.Error(w, "Error while processing request.", http.StatusBadRequest)
	}

}

