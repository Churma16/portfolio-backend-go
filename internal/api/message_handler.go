package api

import (
	"fmt"
	db "go-portfolio-api/db/sqlc"
	"go-portfolio-api/internal/response"
	"go-portfolio-api/internal/util"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

type createMessageRequest struct {
	Name    string `form:"name" binding:"required"`
	Email   string `form:"email" binding:"required,email"`
	Content string `form:"content" binding:"required"`

	// HONEYPOT FIELD
	Gotcha string `form:"gotcha"`
}

func (server *Server) createMessage(ctx *gin.Context) {
	var req createMessageRequest
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 1. LOGIKA HONEYPOT (Sama persis kayak Laravel $request->filled('gotcha'))
	if req.Gotcha != "" {
		go DispatchTelemetry(os.Getenv("MONITOR_URL"), TelemetryEvent{
			Source: "go",
			Type:   "spam_blocked",
			Metadata: map[string]any{
				"reason": "honeypot_triggered",
			},
		})
		meta := response.NewMeta(http.StatusOK, "success", "Pesan terkirim!")
		ctx.JSON(http.StatusOK, response.NewSingleDataResponse(meta, nil))
		return
	}

	// 2. Simpan ke Database
	arg := db.CreateMessageParams{
		Name:    req.Name,
		Email:   req.Email,
		Content: req.Content,
	}

	message, err := server.store.CreateMessage(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, response.ErrorResponse(http.StatusInternalServerError, "error", err.Error()))
		return
	}

	// 3. Kirim Notifikasi via portfolio-notifier (Asynchronous with bounded timeout)
	go func(name, email, content string) {
		if err := util.SendNotification(name, email, content, "go"); err != nil {
			fmt.Printf("[fail] GAGAL KIRIM NOTIFIKASI KE NOTIFIER: %v\n", err)
		} else {
			fmt.Printf("[success] NOTIFIKASI TERKIRIM KE PORTFOLIO-NOTIFIER UNTUK: %s\n", email)
		}
		DispatchTelemetry(os.Getenv("MONITOR_URL"), TelemetryEvent{
			Source: "go",
			Type:   "contact_submit",
			Metadata: map[string]any{
				"email": email,
			},
		})
	}(req.Name, req.Email, req.Content)

	// 4. Return Response
	meta := response.NewMeta(http.StatusOK, "success", "Pesan sedang dikirim!")
	ctx.JSON(http.StatusOK, response.NewSingleDataResponse(meta, message))
}
