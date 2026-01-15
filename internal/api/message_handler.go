package api

import (
	"fmt"
	db "go-portfolio-api/db/sqlc"
	"go-portfolio-api/internal/response"
	"go-portfolio-api/internal/util"
	"net/http"

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
	println(ctx)
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
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 3. Kirim Email (Pake Goroutine biar Cepat!) 🚀
	// Di Laravel ini butuh Queue Worker, di Go cukup "go func()"
	go func() {
		emailData := util.EmailRequest{
			SenderName:  req.Name,
			SenderEmail: req.Email,
			Content:     req.Content,
		}
		// Kita abaikan error di background process untuk kesederhanaan,
		// tapi idealnya di-log.
		err := util.SendContactEmail(emailData)
		if err != nil {
			// Kalau Gagal, print ke terminal server
			fmt.Printf("❌ GAGAL KIRIM EMAIL: %v\n", err)
		} else {
			// Kalau Sukses
			fmt.Printf("✅ EMAIL TERKIRIM KE: %s\n", emailData.SenderEmail)
		}
	}()

	// 4. Return Response
	meta := response.NewMeta(http.StatusOK, "success", "Pesan sedang dikirim!")
	ctx.JSON(http.StatusOK, response.NewSingleDataResponse(meta, message))
}
