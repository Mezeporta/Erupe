package channelserver

import (
	"time"

	"erupe-ce/common/byteframe"
	ps "erupe-ce/common/pascalstring"
	"erupe-ce/common/stringsupport"
	"erupe-ce/network/mhfpacket"
	"go.uber.org/zap"
)

// MessageBoardPost represents a guild message board post.
type MessageBoardPost struct {
	ID        uint32    `db:"id"`
	StampID   uint32    `db:"stamp_id"`
	Title     string    `db:"title"`
	Body      string    `db:"body"`
	AuthorID  uint32    `db:"author_id"`
	Timestamp time.Time `db:"created_at"`
	LikedBy   string    `db:"liked_by"`
}

func handleMsgMhfEnumerateGuildMessageBoard(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateGuildMessageBoard)
	guild, _ := s.server.guildRepo.GetByCharID(s.charID)
	if pkt.BoardType == 1 {
		pkt.MaxPosts = 4
	}
	msgs, err := s.server.db.Queryx("SELECT id, stamp_id, title, body, author_id, created_at, liked_by FROM guild_posts WHERE guild_id = $1 AND post_type = $2 AND deleted = false ORDER BY created_at DESC", guild.ID, int(pkt.BoardType))
	if err != nil {
		s.logger.Error("Failed to get guild messages from db", zap.Error(err))
		doAckBufSucceed(s, pkt.AckHandle, make([]byte, 4))
		return
	}
	if err := s.server.charRepo.UpdateGuildPostChecked(s.charID); err != nil {
		s.logger.Error("Failed to update guild post checked time", zap.Error(err))
	}
	bf := byteframe.NewByteFrame()
	var postCount uint32
	for msgs.Next() {
		postData := &MessageBoardPost{}
		err = msgs.StructScan(&postData)
		if err != nil {
			continue
		}
		postCount++
		bf.WriteUint32(postData.ID)
		bf.WriteUint32(postData.AuthorID)
		bf.WriteUint32(0)
		bf.WriteUint32(uint32(postData.Timestamp.Unix()))
		bf.WriteUint32(uint32(stringsupport.CSVLength(postData.LikedBy)))
		bf.WriteBool(stringsupport.CSVContains(postData.LikedBy, int(s.charID)))
		bf.WriteUint32(postData.StampID)
		ps.Uint32(bf, postData.Title, true)
		ps.Uint32(bf, postData.Body, true)
	}
	data := byteframe.NewByteFrame()
	data.WriteUint32(postCount)
	data.WriteBytes(bf.Data())
	doAckBufSucceed(s, pkt.AckHandle, data.Data())
}

func handleMsgMhfUpdateGuildMessageBoard(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfUpdateGuildMessageBoard)
	guild, err := s.server.guildRepo.GetByCharID(s.charID)
	applicant := false
	if guild != nil {
		applicant, _ = s.server.guildRepo.HasApplication(guild.ID, s.charID)
	}
	if err != nil || guild == nil || applicant {
		doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
		return
	}
	switch pkt.MessageOp {
	case 0: // Create message
		if _, err := s.server.db.Exec("INSERT INTO guild_posts (guild_id, author_id, stamp_id, post_type, title, body) VALUES ($1, $2, $3, $4, $5, $6)", guild.ID, s.charID, pkt.StampID, pkt.PostType, pkt.Title, pkt.Body); err != nil {
			s.logger.Error("Failed to insert guild post", zap.Error(err))
		}
		maxPosts := 100
		if pkt.PostType == 1 {
			maxPosts = 4
		}
		if _, err := s.server.db.Exec(`UPDATE guild_posts SET deleted = true WHERE id IN (
			SELECT id FROM guild_posts WHERE guild_id = $1 AND post_type = $2 AND deleted = false
			ORDER BY created_at DESC OFFSET $3
		)`, guild.ID, pkt.PostType, maxPosts); err != nil {
			s.logger.Error("Failed to soft-delete excess guild posts", zap.Error(err))
		}
	case 1: // Delete message
		if _, err := s.server.db.Exec("UPDATE guild_posts SET deleted = true WHERE id = $1", pkt.PostID); err != nil {
			s.logger.Error("Failed to soft-delete guild post", zap.Error(err))
		}
	case 2: // Update message
		if _, err := s.server.db.Exec("UPDATE guild_posts SET title = $1, body = $2 WHERE id = $3", pkt.Title, pkt.Body, pkt.PostID); err != nil {
			s.logger.Error("Failed to update guild post", zap.Error(err))
		}
	case 3: // Update stamp
		if _, err := s.server.db.Exec("UPDATE guild_posts SET stamp_id = $1 WHERE id = $2", pkt.StampID, pkt.PostID); err != nil {
			s.logger.Error("Failed to update guild post stamp", zap.Error(err))
		}
	case 4: // Like message
		var likedBy string
		err := s.server.db.QueryRow("SELECT liked_by FROM guild_posts WHERE id = $1", pkt.PostID).Scan(&likedBy)
		if err != nil {
			s.logger.Error("Failed to get guild message like data from db", zap.Error(err))
		} else {
			if pkt.LikeState {
				likedBy = stringsupport.CSVAdd(likedBy, int(s.charID))
				if _, err := s.server.db.Exec("UPDATE guild_posts SET liked_by = $1 WHERE id = $2", likedBy, pkt.PostID); err != nil {
					s.logger.Error("Failed to update guild post likes", zap.Error(err))
				}
			} else {
				likedBy = stringsupport.CSVRemove(likedBy, int(s.charID))
				if _, err := s.server.db.Exec("UPDATE guild_posts SET liked_by = $1 WHERE id = $2", likedBy, pkt.PostID); err != nil {
					s.logger.Error("Failed to update guild post likes", zap.Error(err))
				}
			}
		}
	case 5: // Check for new messages
		var newPosts int
		timeChecked, err := s.server.charRepo.ReadGuildPostChecked(s.charID)
		if err == nil {
			_ = s.server.db.QueryRow("SELECT COUNT(*) FROM guild_posts WHERE guild_id = $1 AND deleted = false AND (EXTRACT(epoch FROM created_at)::int) > $2", guild.ID, timeChecked.Unix()).Scan(&newPosts)
			if newPosts > 0 {
				doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x01})
				return
			}
		}
	}
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}
