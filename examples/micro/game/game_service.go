package game

import (
	"log"

	"github.com/lonng/nano/component"
	"github.com/lonng/nano/session"
)

type GameService struct{ component.Base }

func newGameService() *GameService { return &GameService{} }

type HelloRequest struct {
	Name string `json:"name"`
}
type HelloResponse struct {
	Message string `json:"message"`
}

// Hello replies to the client with a Hello World message via session.Response
func (gs *GameService) Hello(s *session.Session, req *HelloRequest) error {
	// log request and respond to the original request (will be routed back to Gate -> client)
	log.Printf("GameService.Hello received: session=%d name=%s", s.ID(), req.Name)
	// explicit marker to ensure we can find the log in aggregated files
	log.Printf("[GAME] Handler Enter: session=%d route=GameService.Hello name=%s", s.ID(), req.Name)
	resp := &HelloResponse{Message: "Hello World, " + req.Name}
	if err := s.Response(resp); err != nil {
		log.Printf("GameService.Hello response error: %v", err)
		return err
	}
	log.Printf("GameService.Hello responded to session=%d: %s", s.ID(), resp.Message)
	log.Printf("[GAME] Handler Exit: session=%d responded=%s", s.ID(), resp.Message)
	return nil
}
