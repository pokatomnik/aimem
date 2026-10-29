package server

import (
	"errors"
	"strings"
)

const (
	AuthorizationHeader = "Authorization"
	bearerMark          = "Bearer"
)

var parseError = errors.New("incorrect header")

func ParseAuthHeader(headerValue string) (string, error) {
	hValSplit := strings.Split(headerValue, " ")
	if len(hValSplit) != 2 || !strings.EqualFold(hValSplit[0], bearerMark) {
		return "", parseError
	}

	return hValSplit[1], nil
}
