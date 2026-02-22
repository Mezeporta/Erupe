package signserver

import (
	"database/sql"
	"errors"
	"erupe-ce/common/mhfcourse"
	"erupe-ce/common/token"
	"time"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

func (s *Server) newUserChara(uid uint32) error {
	numNewChars, err := s.charRepo.CountNewCharacters(uid)
	if err != nil {
		return err
	}

	// prevent users with an uninitialised character from creating more
	if numNewChars >= 1 {
		return nil
	}

	return s.charRepo.CreateCharacter(uid, uint32(time.Now().Unix()))
}

func (s *Server) registerDBAccount(username string, password string) (uint32, error) {
	s.logger.Info("Creating user", zap.String("User", username))

	// Create salted hash of user password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}

	uid, err := s.userRepo.Register(username, string(passwordHash), time.Now().Add(time.Hour*24*30))
	if err != nil {
		return 0, err
	}

	return uid, nil
}

func (s *Server) getCharactersForUser(uid uint32) ([]character, error) {
	return s.charRepo.GetForUser(uid)
}

func (s *Server) getReturnExpiry(uid uint32) time.Time {
	var returnExpiry time.Time
	lastLogin, err := s.userRepo.GetLastLogin(uid)
	if err != nil {
		s.logger.Warn("Failed to get last login", zap.Uint32("uid", uid), zap.Error(err))
		lastLogin = time.Now()
	}
	if time.Now().Add((time.Hour * 24) * -90).After(lastLogin) {
		returnExpiry = time.Now().Add(time.Hour * 24 * 30)
		if err := s.userRepo.UpdateReturnExpiry(uid, returnExpiry); err != nil {
			s.logger.Warn("Failed to update return expiry", zap.Uint32("uid", uid), zap.Error(err))
		}
	} else {
		returnExpiry, err = s.userRepo.GetReturnExpiry(uid)
		if err != nil {
			returnExpiry = time.Now()
			if err := s.userRepo.UpdateReturnExpiry(uid, returnExpiry); err != nil {
				s.logger.Warn("Failed to update return expiry (fallback)", zap.Uint32("uid", uid), zap.Error(err))
			}
		}
	}
	if err := s.userRepo.UpdateLastLogin(uid, time.Now()); err != nil {
		s.logger.Warn("Failed to update last login", zap.Uint32("uid", uid), zap.Error(err))
	}
	return returnExpiry
}

func (s *Server) getLastCID(uid uint32) uint32 {
	lastPlayed, err := s.userRepo.GetLastCharacter(uid)
	if err != nil {
		s.logger.Warn("Failed to get last character", zap.Uint32("uid", uid), zap.Error(err))
		return 0
	}
	return lastPlayed
}

func (s *Server) getUserRights(uid uint32) uint32 {
	if uid == 0 {
		return 0
	}
	rights, err := s.userRepo.GetRights(uid)
	if err != nil {
		s.logger.Warn("Failed to get user rights", zap.Uint32("uid", uid), zap.Error(err))
		return 0
	}
	_, rights = mhfcourse.GetCourseStruct(rights, s.erupeConfig.DefaultCourses)
	return rights
}

func (s *Server) getFriendsForCharacters(chars []character) []members {
	friends := make([]members, 0)
	for _, char := range chars {
		charFriends, err := s.charRepo.GetFriends(char.ID)
		if err != nil {
			s.logger.Warn("Failed to get friends", zap.Uint32("charID", char.ID), zap.Error(err))
			continue
		}
		for i := range charFriends {
			charFriends[i].CID = char.ID
		}
		friends = append(friends, charFriends...)
	}
	return friends
}

func (s *Server) getGuildmatesForCharacters(chars []character) []members {
	guildmates := make([]members, 0)
	for _, char := range chars {
		charGuildmates, err := s.charRepo.GetGuildmates(char.ID)
		if err != nil {
			s.logger.Warn("Failed to get guildmates", zap.Uint32("charID", char.ID), zap.Error(err))
			continue
		}
		for i := range charGuildmates {
			charGuildmates[i].CID = char.ID
		}
		guildmates = append(guildmates, charGuildmates...)
	}
	return guildmates
}

func (s *Server) deleteCharacter(cid int, tok string, tokenID uint32) error {
	if !s.validateToken(tok, tokenID) {
		return errors.New("invalid token")
	}
	isNew, err := s.charRepo.IsNewCharacter(cid)
	if err != nil {
		return err
	}
	if isNew {
		return s.charRepo.HardDelete(cid)
	}
	return s.charRepo.SoftDelete(cid)
}

func (s *Server) registerUidToken(uid uint32) (uint32, string, error) {
	_token := token.Generate(16)
	tid, err := s.sessionRepo.RegisterUID(uid, _token)
	return tid, _token, err
}

func (s *Server) registerPsnToken(psn string) (uint32, string, error) {
	_token := token.Generate(16)
	tid, err := s.sessionRepo.RegisterPSN(psn, _token)
	return tid, _token, err
}

func (s *Server) validateToken(tok string, tokenID uint32) bool {
	valid, err := s.sessionRepo.Validate(tok, tokenID)
	if err != nil {
		s.logger.Warn("Failed to validate token", zap.Error(err))
		return false
	}
	return valid
}

func (s *Server) validateLogin(user string, pass string) (uint32, RespID) {
	uid, passDB, err := s.userRepo.GetCredentials(user)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.logger.Info("User not found", zap.String("User", user))
			if s.erupeConfig.AutoCreateAccount {
				uid, err = s.registerDBAccount(user, pass)
				if err == nil {
					return uid, SIGN_SUCCESS
				}
				return 0, SIGN_EABORT
			}
			return 0, SIGN_EAUTH
		}
		return 0, SIGN_EABORT
	}

	if bcrypt.CompareHashAndPassword([]byte(passDB), []byte(pass)) != nil {
		return 0, SIGN_EPASS
	}

	bans, err := s.userRepo.CountPermanentBans(uid)
	if err == nil && bans > 0 {
		return uid, SIGN_EELIMINATE
	}
	bans, err = s.userRepo.CountActiveBans(uid)
	if err == nil && bans > 0 {
		return uid, SIGN_ESUSPEND
	}
	return uid, SIGN_SUCCESS
}
