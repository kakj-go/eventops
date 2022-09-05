/*
 * Copyright 2022 The kakj-go Authors.
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *     http://www.apache.org/licenses/LICENSE-2.0
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package token

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"time"
)

type IssuerType string

const Login IssuerType = "login"

func (s IssuerType) String() string {
	return string(s)
}

type SubjectType string

const EventOps SubjectType = "eventops"

func (s SubjectType) String() string {
	return string(s)
}

func GenLoginToken(name, signature string, expiresAt time.Time) (string, error) {
	// iss：发行人
	// exp：到期时间
	// sub：主题
	// aud：用户
	// nbf：在此之前不可用
	// iat：发布时间
	// jti：JWT ID用于标识该JWT
	claims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(expiresAt),
		Subject:   Login.String(),
		Issuer:    EventOps.String(),
		Audience:  jwt.ClaimStrings{name},
		NotBefore: jwt.NewNumericDate(time.Now()),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString([]byte(signature))
	if err != nil {
		return "", err
	}
	return ss, err
}

func ParseLoginToken(token string, signature string) (jwtToken *jwt.Token, audience string, err error) {
	parseToken, err := jwt.ParseWithClaims(token, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(signature), nil
	})
	if err != nil {
		return nil, "", err
	}

	if claims, ok := parseToken.Claims.(*jwt.RegisteredClaims); ok && parseToken.Valid {
		return parseToken, claims.Audience[0], nil
	} else {
		return nil, "", fmt.Errorf("token valid failed")
	}
}

func UpdateExpiresTime(token *jwt.Token, expiresAt time.Time, signature string) (string, error) {
	if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok && token.Valid {
		claims.ExpiresAt = jwt.NewNumericDate(expiresAt)
	}

	tokenString, err := token.SignedString([]byte(signature))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func SetToken(c *gin.Context, tokenValue string) {
	c.Header(AuthHeader, BuildTokenHeaderValue(tokenValue))
}

func GetUserName(c *gin.Context) string {
	return c.GetHeader(UserNameHeader)
}
