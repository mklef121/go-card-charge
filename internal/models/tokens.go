package models

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"time"
)

const (
	ScopeAuthentication = "Authentication"
)

type Token struct {
	PlainText string    `json:"token"`
	UserID    int64     `json:"-"`
	Hash      []byte    `json:"-"`
	Expiry    time.Time `json:"expiry"`
	Scope     string    `json:"-"`
}

//Generate a token that lasts for the expected time
func GenerateToken(UserID int, ttl time.Duration, scope string) (*Token, error) {
	token := Token{
		UserID: int64(UserID),
		Scope:  scope,
		Expiry: time.Now().Add(ttl),
	}

	//Generate a random byte of number (total length will be 16)
	randomByte := make([]byte, 16)

	_, err := rand.Read(randomByte)

	if err != nil {
		return nil, err
	}

	//Convert the byte array (numbers) to JSON encoded text
	token.PlainText = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomByte)

	hash := sha256.Sum256([]byte(token.PlainText))
	token.Hash = hash[:]

	return &token, nil
}

func (model *DBModel) InsertToken(token *Token, user User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stm := `insert into tokens (user_id,name,email,token_hash) 
				values(?,?,?,?)`

	_, err := model.DB.ExecContext(ctx,
		stm,
		user.ID,
		user.LastName,
		user.Email,
		token.Hash,
	)

	if err != nil {
		return err
	}

	return nil
}
