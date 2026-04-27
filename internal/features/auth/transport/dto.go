package transport

import (
	core_domain "GopherDrive/internal/core/domain"
	"GopherDrive/internal/pkg/api/response"
	"net/http"
)

type Request struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserResponse struct {
	response.Response
	AccessToken string           `json:"access_token"`
	User        core_domain.User `json:"user"`
}

type RefreshTokenResponse struct {
	response.Response
	AccessToken string `json:"access_token"`
}

func CreateUserResponse(response response.Response, accessToken string, user core_domain.User) *UserResponse {
	return &UserResponse{
		Response:    response,
		AccessToken: accessToken,
		User:        user,
	}
}

func CreateRefreshTokenResponse(response response.Response, accessToken string) *RefreshTokenResponse {
	return &RefreshTokenResponse{
		Response:    response,
		AccessToken: accessToken,
	}
}

func createCookie(name string, value string, ttl int) *http.Cookie {
	return &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		MaxAge:   ttl,
		Secure:   false,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
}
