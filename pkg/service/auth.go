package service

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/Frol333/14Sprint/pkg/config"
	"github.com/Frol333/14Sprint/pkg/db"
	"golang.org/x/crypto/bcrypt"
)

func CheckPassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if config.Password != req.Password {
		http.Error(w, "password is not correct", http.StatusUnauthorized)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(config.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to get tasks", http.StatusUnauthorized)
		return
	}

	err = db.SavePasswordHash(r.Context(), string(hash))
	if err != nil {
		http.Error(w, "Failed to save password hash", http.StatusInternalServerError)
		return
	}

	// 5. отправляем JSON
	writeJson(w, map[string]string{
		"token": string(hash),
	})
}

func CheckAuth(ctx context.Context, key string) (bool, error) {
	_, err := db.GetUserIDByPasswordHash(ctx, key)
	if err != nil {
		return false, err
	}

	return true, nil

}
