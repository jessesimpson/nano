package gate

import (
	"github.com/lonng/nano/component"
	"github.com/lonng/nano/session"
)

type GateService struct{ component.Base }

func newGateService() *GateService { return &GateService{} }

type EchoRequest struct {
	Name string `json:"name"`
}

// Hello forwards the request to GameService.Hello on game nodes.
func (gs *GateService) Hello(s *session.Session, msg *EchoRequest) error {
	// log receipt and forward to GameService; response will be routed back to client
	return nil
}
