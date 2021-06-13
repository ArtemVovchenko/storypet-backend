package auth

import (
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type IoTAccessTokenMeta struct {
	PetID   int
	Expires int64
	Token   string
}

func CreateIoTToken(petID int) (*IoTAccessTokenMeta, error) {
	tInfo := &IoTAccessTokenMeta{
		PetID:   petID,
		Expires: time.Now().Add(time.Hour * 24).Unix(),
	}

	accessClaims := jwt.MapClaims{}
	accessClaims["pet_id"] = petID
	accessClaims["expires_at"] = tInfo.Expires
	rawAccessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessToken, err := rawAccessToken.SignedString([]byte(os.Getenv("ACCESS_SECRET")))
	if err != nil {
		return nil, err
	}
	tInfo.Token = accessToken
	return tInfo, err
}

func extractIoTToken(r *http.Request) (string, error) {
	bearToken := r.Header.Get("Authorization")
	authData := strings.Split(bearToken, " ")
	if len(authData) == 2 {
		return authData[1], nil
	}
	return "", errors.New("request's authorization field is unprocessable")
}

func verifyIoTAccessToken(r *http.Request) (*jwt.Token, error) {
	tokenStr, err := extractIoTToken(r)
	if err != nil {
		return nil, err
	}

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("token singing method is unverified")
		}
		return []byte(os.Getenv("ACCESS_SECRET")), nil
	})
	if err != nil {
		return nil, err
	}
	return token, nil
}

func ExtractIoTAccessMeta(r *http.Request) (*IoTAccessTokenMeta, error) {
	token, err := verifyIoTAccessToken(r)
	if err != nil {
		return nil, err
	}

	accessClaims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		petID, err := strconv.ParseInt(fmt.Sprintf("%v", accessClaims["pet_id"]), 10, 64)
		if err != nil {
			return nil, err
		}
		expiresAt, err := strconv.ParseUint(fmt.Sprintf("%.f", accessClaims["expires_at"]), 10, 64)
		if err != nil {
			return nil, err
		}
		return &IoTAccessTokenMeta{
			PetID:   int(petID),
			Expires: int64(expiresAt),
		}, nil
	}
	return nil, err
}
