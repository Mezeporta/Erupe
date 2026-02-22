package api

import (
	"context"
	"database/sql"
	"erupe-ce/common/token"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func (s *APIServer) createNewUser(ctx context.Context, username string, password string) (uint32, uint32, error) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, 0, err
	}
	return s.userRepo.Register(ctx, username, string(passwordHash), time.Now().Add(time.Hour*24*30))
}

func (s *APIServer) createLoginToken(ctx context.Context, uid uint32) (uint32, string, error) {
	loginToken := token.Generate(16)
	tid, err := s.sessionRepo.CreateToken(ctx, uid, loginToken)
	if err != nil {
		return 0, "", err
	}
	return tid, loginToken, nil
}

func (s *APIServer) userIDFromToken(ctx context.Context, tkn string) (uint32, error) {
	userID, err := s.sessionRepo.GetUserIDByToken(ctx, tkn)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, fmt.Errorf("invalid login token")
	} else if err != nil {
		return 0, err
	}
	return userID, nil
}

func (s *APIServer) createCharacter(ctx context.Context, userID uint32) (Character, error) {
	character, err := s.charRepo.GetNewCharacter(ctx, userID)
	if errors.Is(err, sql.ErrNoRows) {
		count, _ := s.charRepo.CountForUser(ctx, userID)
		if count >= 16 {
			return character, fmt.Errorf("cannot have more than 16 characters")
		}
		character, err = s.charRepo.Create(ctx, userID, uint32(time.Now().Unix()))
	}
	return character, err
}

func (s *APIServer) deleteCharacter(_ context.Context, _ uint32, charID uint32) error {
	isNew, err := s.charRepo.IsNew(charID)
	if err != nil {
		return err
	}
	if isNew {
		return s.charRepo.HardDelete(charID)
	}
	return s.charRepo.SoftDelete(charID)
}

func (s *APIServer) getCharactersForUser(ctx context.Context, uid uint32) ([]Character, error) {
	return s.charRepo.GetForUser(ctx, uid)
}

func (s *APIServer) getReturnExpiry(uid uint32) time.Time {
	lastLogin, _ := s.userRepo.GetLastLogin(uid)
	var returnExpiry time.Time
	if time.Now().Add((time.Hour * 24) * -90).After(lastLogin) {
		returnExpiry = time.Now().Add(time.Hour * 24 * 30)
		_ = s.userRepo.UpdateReturnExpiry(uid, returnExpiry)
	} else {
		var err error
		returnExpiry, err = s.userRepo.GetReturnExpiry(uid)
		if err != nil {
			returnExpiry = time.Now()
			_ = s.userRepo.UpdateReturnExpiry(uid, returnExpiry)
		}
	}
	_ = s.userRepo.UpdateLastLogin(uid, time.Now())
	return returnExpiry
}

func (s *APIServer) exportSave(ctx context.Context, uid uint32, cid uint32) (map[string]interface{}, error) {
	return s.charRepo.ExportSave(ctx, uid, cid)
}
