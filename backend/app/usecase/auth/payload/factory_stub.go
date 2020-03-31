package payload

import (
	"github.com/short-d/app/fw"
	"github.com/short-d/short/app/entity"
)

var _ Factory = (*FactoryStub)(nil)

type FactoryStub struct {
	Payload  Payload
	TokenErr error
	UserErr  error
}

func (f FactoryStub) FromTokenPayload(tokenPayload fw.TokenPayload) (Payload, error) {
	return f.Payload, f.TokenErr
}

func (f FactoryStub) FromUser(user entity.User) (Payload, error) {
	return f.Payload, f.UserErr
}
