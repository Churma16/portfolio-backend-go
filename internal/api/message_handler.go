package api

import (
	"fmt"
	db "go-portfolio-api/db/sqlc"
	"go-portfolio-api/internal/response"
	"go-portfolio-api/internal/util"
	"net/http"
	"time"

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

	// 3. Kirim Email (Asynchronous with bounded timeout)
	emailData := util.EmailRequest{
		SenderName:  req.Name,
		SenderEmail: req.Email,
		Content:     req.Content,
	}

	go func(data util.EmailRequest) {
		done := make(chan error, 1)
		go func() {
			done <- util.SendContactEmail(data)
		}()

		select {
		case err := <-done:
			if err != nil {
				fmt.Printf("[fail] GAGAL KIRIM EMAIL: %v\n", err)
			} else {
				fmt.Printf("[success] EMAIL TERKIRIM KE: %s\n", data.SenderEmail)
			}
		case <-time.After(15 * time.Second):
			fmt.Printf("[fail] TIMEOUT KIRIM EMAIL KE: %s\n", data.SenderEmail)
		}
	}(emailData)

	// 4. Return Response
	meta := response.NewMeta(http.StatusOK, "success", "Pesan sedang dikirim!")
	ctx.JSON(http.StatusOK, response.NewSingleDataResponse(meta, message))
}
