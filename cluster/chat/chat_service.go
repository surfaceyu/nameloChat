package chat

import (
	"fmt"
	"log"

	"github.com/pingcap/errors"
	"github.com/surfaceyu/namelo"
	"github.com/surfaceyu/namelo/component"
	"github.com/surfaceyu/namelo/examples/cluster/protocol"
	"github.com/surfaceyu/namelo/session"
)

type RoomService struct {
	component.Base
	group *namelo.Group
}

func newRoomService() *RoomService {
	return &RoomService{
		group: namelo.NewGroup("all-users"),
	}
}

func (rs *RoomService) JoinRoom(s *session.Session, msg *protocol.JoinRoomRequest) error {
	if err := s.Bind(msg.MasterUid); err != nil {
		return errors.Trace(err)
	}

	broadcast := &protocol.NewUserBroadcast{
		Content: fmt.Sprintf("User user join: %v", msg.Nickname),
	}
	if err := rs.group.Broadcast("onNewUser", broadcast); err != nil {
		return errors.Trace(err)
	}
	return rs.group.Add(s)
}

type SyncMessage struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

func (rs *RoomService) SyncMessage(s *session.Session, msg *SyncMessage) error {
	// Send an RPC to master server to stats
	if err := s.RPC("TopicService.Stats", &protocol.MasterStats{Uid: s.UID()}); err != nil {
		return errors.Trace(err)
	}

	// Sync message to all members in this room
	return rs.group.Broadcast("onMessage", msg)
}

func (rs *RoomService) userDisconnected(s *session.Session) {
	if err := rs.group.Leave(s); err != nil {
		log.Println("Remove user from group failed", s.UID(), err)
		return
	}
	log.Println("User session disconnected", s.UID())
}
