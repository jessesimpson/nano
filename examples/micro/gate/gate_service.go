package gate

import (
	"log"

	"github.com/lonng/nano/component"
	"github.com/lonng/nano/session"
)

type GateService struct{ component.Base }

func newGateService() *GateService { return &GateService{} }

type HelloRequest struct {
	Name string `json:"name"`
}

// Hello forwards the request to GameService.Hello on game nodes.
func (gs *GateService) Hello(s *session.Session, msg *HelloRequest) error {
	// log receipt and forward to GameService; response will be routed back to client
	log.Printf("GateService.Hello received: session=%d name=%s", s.ID(), msg.Name)
	if err := s.RPC("GameService.Hello", msg); err != nil {
		log.Printf("GateService.Hello RPC error: %v", err)
		return err
	}
	log.Printf("GateService.Hello forwarded to GameService for session=%d", s.ID())
	return nil
}
