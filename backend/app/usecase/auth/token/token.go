package token

import (
	"time"

	"github.com/short-d/short/app/entity"
)

type Token struct {
	User     entity.User
	IssuedAt time.Time
}
