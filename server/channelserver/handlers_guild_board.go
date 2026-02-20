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
	posts, err := s.server.guildRepo.ListPosts(guild.ID, int(pkt.BoardType))
	if err != nil {
		s.logger.Error("Failed to get guild messages from db", zap.Error(err))
		doAckBufSucceed(s, pkt.AckHandle, make([]byte, 4))
		return
	}
	if err := s.server.charRepo.UpdateGuildPostChecked(s.charID); err != nil {
		s.logger.Error("Failed to update guild post checked time", zap.Error(err))
	}
	bf := byteframe.NewByteFrame()
	for _, postData := range posts {
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
	data.WriteUint32(uint32(len(posts)))
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
		maxPosts := 100
		if pkt.PostType == 1 {
			maxPosts = 4
		}
		if err := s.server.guildRepo.CreatePost(guild.ID, s.charID, pkt.StampID, int(pkt.PostType), pkt.Title, pkt.Body, maxPosts); err != nil {
			s.logger.Error("Failed to create guild post", zap.Error(err))
		}
	case 1: // Delete message
		if err := s.server.guildRepo.DeletePost(pkt.PostID); err != nil {
			s.logger.Error("Failed to soft-delete guild post", zap.Error(err))
		}
	case 2: // Update message
		if err := s.server.guildRepo.UpdatePost(pkt.PostID, pkt.Title, pkt.Body); err != nil {
			s.logger.Error("Failed to update guild post", zap.Error(err))
		}
	case 3: // Update stamp
		if err := s.server.guildRepo.UpdatePostStamp(pkt.PostID, pkt.StampID); err != nil {
			s.logger.Error("Failed to update guild post stamp", zap.Error(err))
		}
	case 4: // Like message
		likedBy, err := s.server.guildRepo.GetPostLikedBy(pkt.PostID)
		if err != nil {
			s.logger.Error("Failed to get guild message like data from db", zap.Error(err))
		} else {
			if pkt.LikeState {
				likedBy = stringsupport.CSVAdd(likedBy, int(s.charID))
			} else {
				likedBy = stringsupport.CSVRemove(likedBy, int(s.charID))
			}
			if err := s.server.guildRepo.SetPostLikedBy(pkt.PostID, likedBy); err != nil {
				s.logger.Error("Failed to update guild post likes", zap.Error(err))
			}
		}
	case 5: // Check for new messages
		timeChecked, err := s.server.charRepo.ReadGuildPostChecked(s.charID)
		if err == nil {
			newPosts, _ := s.server.guildRepo.CountNewPosts(guild.ID, timeChecked)
			if newPosts > 0 {
				doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x01})
				return
			}
		}
	}
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}
