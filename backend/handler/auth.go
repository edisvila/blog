package handler

import (
	"context"
	"os"
	"time"

	"blog_backend/api"
	"blog_backend/model"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func (h *Handler) PostAuthLogin(ctx context.Context, request api.PostAuthLoginRequestObject) (api.PostAuthLoginResponseObject, error) {
	var user model.User
	result := h.db.Where("username = ?", request.Body.Username).First(&user)
	if result.Error != nil {
		return api.PostAuthLogin401Response{}, nil
	}

	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(request.Body.Password))
	if err != nil {
		return api.PostAuthLogin401Response{}, nil
	}

	token, err := generateToken(user.ID)
	if err != nil {
		return nil, err
	}

	return api.PostAuthLogin200JSONResponse(api.TokenResponse{
		Token: &token,
	}), nil
}

func generateToken(userID uint) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}
