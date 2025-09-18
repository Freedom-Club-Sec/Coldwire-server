package httpserver

import (
    "log/slog"
    "net/http"
    "encoding/json"
    "io"
    "time"

    "github.com/golang-jwt/jwt/v5"
    "github.com/Freedom-Club-Sec/Coldwire-server/internal/types"
    "github.com/Freedom-Club-Sec/Coldwire-server/internal/constants"
)





func (s *Server) newDataHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    ctx := r.Context()
    jwtClaims := ctx.Value(claimsKey).(jwt.MapClaims)
    userId := jwtClaims["user_id"].(string)


    err := r.ParseMultipartForm(3 << 20) // 3 MB max memory
    if err != nil {
        http.Error(w, "Failed to parse form: "+err.Error(), http.StatusBadRequest)
        return
    }

    metadataStr := r.FormValue("metadata")
    if metadataStr == "" {
        http.Error(w, "Missing metadata", http.StatusBadRequest)
        return
    }
    
    var metadata types.DataSendRequest 

    if err := json.Unmarshal([]byte(metadataStr), &metadata); err != nil {
        slog.Error("Error while parsing request JSON metadata.", "error", err)
        http.Error(w, "Invalid JSON metadata.", http.StatusBadRequest)
        return
    }

    if metadata.Recipient == "" {
        slog.Error("Empty recipient from request metadata.", "metadata", metadata)
        http.Error(w, "Missing recipient in metadata", http.StatusBadRequest)
        return
    }


    file, _, err := r.FormFile("blob")
    if err != nil {
        slog.Error("Failed to read blob from form.", "userId", userId, "error", err)
        http.Error(w, "Failed to read blob from form.", http.StatusBadRequest)
        return
    }
    defer file.Close()

    blobData, err := io.ReadAll(file)
    if err != nil {
        slog.Error("Failed to read blob data.", "userId", userId, "error", err)
        http.Error(w, "Failed to read blob data.", http.StatusBadRequest)
        return
    }

    if len(blobData) == 0 {
        http.Error(w, "Empty blob is not allowed", http.StatusBadRequest)
        return
    }
    

    if err := s.DbSvcs.DataService.InsertData(blobData, userId, metadata.Recipient); err != nil {
        slog.Error("Failure when attempted to insert data.", "userId", userId, "error", err, "metadata", metadata, "blobData", blobData)
        http.Error(w, "Failed to process data.", http.StatusBadRequest)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(`{"status":"success"}`))
}

func (s *Server) dataLongpollHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    ctx := r.Context()
    jwtClaims := ctx.Value(claimsKey).(jwt.MapClaims)
    userId := jwtClaims["user_id"].(string)


	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	timeout := time.NewTimer(time.Second * constants.LONGPOLL_MAX)
	defer timeout.Stop()

    slog.Info("Received data longpoll~!!!")

	for {
		select {
		case <-ctx.Done():
			// client disconnected, stop, don't consume data, don't write
			return
		case <-timeout.C:
			w.Header().Set("Content-Type", "application/octet-stream")
			w.WriteHeader(http.StatusOK)
			return
		case <-ticker.C:
			if ctx.Err() != nil {
				return
			}
		     dataBlobs, err := s.DbSvcs.DataService.GetLatestData(userId)
             if err != nil {
                 slog.Error("Error while getting latest data", "userId", userId, "error", err)
                 http.Error(w, "Error while processing request.", http.StatusBadRequest)
                 return
             }

             if dataBlobs != nil {
                 w.Header().Set("Content-Type", "application/octet-stream")
                 w.WriteHeader(http.StatusOK)
                 _, _ = w.Write(dataBlobs)
                 return
             }
        }
	}
}

