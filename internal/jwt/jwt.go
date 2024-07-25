package jwt

import (
	"grpc_service/internal/domain/models"
	"time"

	"github.com/golang-jwt/jwt"
)

// TODO: Тесты
func NewToken(
	user models.User,
	app models.App,
	duration time.Duration,
) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256) //?-Какой выбрать метод

	claims := token.Claims.(jwt.MapClaims)
	claims["userID"] = user.ID
	claims["email"] = user.Email
	claims["ext"] = time.Now().Add(duration).Unix()
	claims["appID"] = app.ID

	//!-Быть аккуратным и не залогировать app.Secret
	tokenString, err := token.SignedString([]byte(app.Secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
