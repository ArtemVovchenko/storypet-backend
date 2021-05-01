package auth

import (
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/twinj/uuid"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type TokenPairInfo struct {
	AccessToken    string
	RefreshToken   string
	AccessUUID     string
	RefreshUUID    string
	AccessExpires  int64
	RefreshExpires int64
}

type AccessTokenMeta struct {
	Authorized bool
	AccessUUID string
	UserID     int
	Expires    int64
}

type RefreshTokenMeta struct {
	RefreshUUID string
	UserID      int
	Expires     int64
}

func CreateToken(userID int) (*TokenPairInfo, error) {
	tInfo := &TokenPairInfo{
		AccessUUID:     uuid.NewV4().String(),
		RefreshUUID:    uuid.NewV4().String(),
		AccessExpires:  time.Now().Add(time.Hour * 24 * 7).Unix(),
		RefreshExpires: time.Now().Add(time.Hour * 24 * 7).Unix(),
	}

	accessClaims := jwt.MapClaims{}
	accessClaims["authorized"] = true
	accessClaims["access_uuid"] = tInfo.AccessUUID
	accessClaims["user_id"] = userID
	accessClaims["expires_at"] = tInfo.AccessExpires
	rawAccessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessToken, err := rawAccessToken.SignedString([]byte(os.Getenv("ACCESS_SECRET")))
	if err != nil {
		return nil, err
	}

	refreshClaims := jwt.MapClaims{}
	refreshClaims["refresh_uuid"] = tInfo.RefreshUUID
	refreshClaims["user_id"] = userID
	refreshClaims["expires_at"] = tInfo.RefreshExpires
	rawRefreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshToken, err := rawRefreshToken.SignedString([]byte(os.Getenv("REFRESH_SECRET")))
	if err != nil {
		return nil, err
	}

	tInfo.AccessToken = accessToken
	tInfo.RefreshToken = refreshToken
	return tInfo, nil
}

func extractToken(r *http.Request) (string, error) {
	bearToken := r.Header.Get("Authorization")
	authData := strings.Split(bearToken, " ")
	if len(authData) == 2 {
		return authData[1], nil
	}
	return "", errors.New("request's authorization field is unprocessable")
}

func verifyAccessToken(r *http.Request) (*jwt.Token, error) {
	tokenStr, err := extractToken(r)
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

func verifyRefreshToken(tokenStr string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("token singing method is unverified")
		}
		return []byte(os.Getenv("REFRESH_SECRET")), nil
	})
	if err != nil {
		return nil, err
	}
	return token, nil
}

func ExtractAccessMeta(r *http.Request) (*AccessTokenMeta, error) {
	token, err := verifyAccessToken(r)
	if err != nil {
		return nil, err
	}

	accessClaims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		authorized, err := strconv.ParseBool(fmt.Sprintf("%v", accessClaims["authorized"]))
		if err != nil {
			return nil, err
		}
		accessUUID, ok := accessClaims["access_uuid"].(string)
		if !ok {
			return nil, errors.New("invalid access_uuid")
		}
		userID, err := strconv.ParseInt(fmt.Sprintf("%v", accessClaims["user_id"]), 10, 64)
		if err != nil {
			return nil, err
		}
		expiresAt, err := strconv.ParseUint(fmt.Sprintf("%.f", accessClaims["expires_at"]), 10, 64)
		if err != nil {
			return nil, err
		}
		return &AccessTokenMeta{
			Authorized: authorized,
			AccessUUID: accessUUID,
			UserID:     int(userID),
			Expires:    int64(expiresAt),
		}, nil
	}
	return nil, err
}

func ExtractRefreshMeta(tokenStr string) (*RefreshTokenMeta, error) {
	token, err := verifyRefreshToken(tokenStr)
	if err != nil {
		return nil, err
	}

	accessClaims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		refreshUUID, ok := accessClaims["refresh_uuid"].(string)
		if !ok {
			return nil, errors.New("invalid refresh_uuid")
		}
		userID, err := strconv.ParseInt(fmt.Sprintf("%v", accessClaims["user_id"]), 10, 64)
		if err != nil {
			return nil, err
		}
		expiresAt, err := strconv.ParseUint(fmt.Sprintf("%.f", accessClaims["expires_at"]), 10, 64)
		if err != nil {
			return nil, err
		}
		return &RefreshTokenMeta{
			RefreshUUID: refreshUUID,
			UserID:      int(userID),
			Expires:     int64(expiresAt),
		}, nil
	}
	return nil, err
}
