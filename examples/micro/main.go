package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/lonng/nano"
	"github.com/lonng/nano/examples/micro/game"
	"github.com/lonng/nano/examples/micro/gate"
	"github.com/lonng/nano/serialize/json"
	"github.com/lonng/nano/session"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "NanoMicroDemo"
	app.Commands = []cli.Command{
		{
			Name:   "master",
			Usage:  "start master (serves demo web)",
			Action: runMaster,
		},
		{
			Name: "gate",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "listen,l", Value: "127.0.0.1:34569"},
				cli.StringFlag{Name: "gate-address", Value: "127.0.0.1:34590"},
				cli.StringFlag{Name: "master", Value: "127.0.0.1:34567"},
			},
			Action: runGate,
		},
		{
			Name: "game",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "listen,l", Value: "127.0.0.1:34568"},
				cli.StringFlag{Name: "master", Value: "127.0.0.1:34567"},
			},
			Action: runGame,
		},
	}
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	if err := app.Run(os.Args); err != nil {
		log.Fatalf("Startup server error %+v", err)
	}
}

func srcPath() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Dir(file)
}

func runMaster(c *cli.Context) error {
	// serve the demo web from existing cluster example web
	webDir := filepath.Join(srcPath(), "..", "cluster", "master", "web")
	log.Println("Micro master serving web from", webDir)
	go func() {
		if err := http.ListenAndServe(":12345", http.FileServer(http.Dir(webDir))); err != nil {
			panic(err)
		}
	}()

	// no components, just run as master for cluster registration
	session.Lifetime.OnClosed(func(s *session.Session) {})

	nano.Listen("127.0.0.1:34567",
		nano.WithMaster(),
		nano.WithSerializer(json.NewSerializer()),
		nano.WithDebugMode(),
	)
	return nil
}

func runGate(c *cli.Context) error {
	listen := c.String("listen")
	gateAddr := c.String("gate-address")
	master := c.String("master")
	nano.Listen(listen,
		nano.WithAdvertiseAddr(master),
		nano.WithClientAddr(gateAddr),
		nano.WithComponents(gate.Services),
		nano.WithSerializer(json.NewSerializer()),
		nano.WithIsWebsocket(true),
		nano.WithWSPath("/nano"),
		nano.WithDebugMode(),
	)
	return nil
}

func runGame(c *cli.Context) error {
	listen := c.String("listen")
	master := c.String("master")
	nano.Listen(listen,
		nano.WithAdvertiseAddr(master),
		nano.WithComponents(game.Services),
		nano.WithSerializer(json.NewSerializer()),
		nano.WithDebugMode(),
	)
	return nil
}
