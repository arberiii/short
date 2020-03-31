package token

import (
	"errors"
	"time"

	"github.com/short-d/app/fw"
	"github.com/short-d/short/app/entity"
	"github.com/short-d/short/app/usecase/auth/payload"
)

const issuedAtKey = "issued_at"

type Issuer struct {
	tokenizer      fw.CryptoTokenizer
	timer          fw.Timer
	payloadFactory payload.Factory
}

func (i Issuer) IssuedToken(user entity.User) (string, error) {
	issuedAt := i.timer.Now()
	p, err := i.payloadFactory.FromUser(user)
	if err != nil {
		return "", err
	}
	tokenPayload := p.GetTokenPayload()
	tokenPayload[issuedAtKey] = issuedAt
	return i.tokenizer.Encode(tokenPayload)
}

func (i Issuer) ParseToken(token string) (Token, error) {
	tokenPayload, err := i.tokenizer.Decode(token)
	if err != nil {
		return Token{}, err
	}
	issuedAtStr, ok := tokenPayload[issuedAtKey]
	if !ok {
		return Token{}, errors.New("token is missing issued_at")
	}
	issuedAt, err := parseTime(issuedAtStr)
	if err != nil {
		return Token{}, err
	}

	p, err := i.payloadFactory.FromTokenPayload(tokenPayload)
	if err != nil {
		return Token{}, err
	}
	user := p.GetUser()
	return Token{
		User:     user,
		IssuedAt: issuedAt,
	}, nil
}

func parseTime(timeAny interface{}) (time.Time, error) {
	timeStr, ok := timeAny.(string)
	if !ok {
		return time.Time{}, errors.New("time is not string")
	}
	return time.Parse(time.RFC3339, timeStr)
}

type IssuerFactory struct {
	tokenizer fw.CryptoTokenizer
	timer     fw.Timer
}

func (i IssuerFactory) MakeIssuer(payloadFactory payload.Factory) Issuer {
	return Issuer{
		tokenizer:      i.tokenizer,
		timer:          i.timer,
		payloadFactory: payloadFactory,
	}
}

func NewIssuerFactory(tokenizer fw.CryptoTokenizer, timer fw.Timer) IssuerFactory {
	return IssuerFactory{
		tokenizer: tokenizer,
		timer:     timer,
	}
}
