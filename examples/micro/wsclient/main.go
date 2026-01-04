package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
	"github.com/lonng/nano/examples/micro/game"
)

// This is a very small client that uses the same simple framing as the example
// web client: it performs a handshake and sends a data package containing a
// message request in plain JSON route form. It does not implement full
// encode/decode optimizations; it's enough to call GateService.Hello and
// print the returned payload.

func main() {
	u := url.URL{Scheme: "ws", Host: "127.0.0.1:34590", Path: "/nano"}
	log.Println("connect", u.String())
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	// perform a minimal handshake using the pomelo package format used by the
	// examples. We'll send a handshake package (type 1) with a small JSON body.
	hb := map[string]interface{}{"sys": map[string]interface{}{"heartbeat": 30}}
	b, _ := json.Marshal(hb)
	pkg := encodePackage(1, b)
	if err := c.WriteMessage(websocket.BinaryMessage, pkg); err != nil {
		log.Fatal(err)
	}

	// read handshake response
	c.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, msg, err := c.ReadMessage()
	if err != nil {
		log.Fatal("read handshake resp:", err)
	}
	fmt.Println("handshake resp len", len(msg))

	// send handshake ACK (packet type 2) so server moves session to working state
	ack := encodePackage(2, nil)
	if err := c.WriteMessage(websocket.BinaryMessage, ack); err != nil {
		log.Fatal("write handshake ack:", err)
	}

	// send a request to GameService.Hello with JSON body using the server's HelloRequest
	// build nano Message binary: Request with id=1, route="GameService.Hello", body using game.HelloRequest
	req := game.HelloRequest{Name: "cli"}
	bb, _ := json.Marshal(req)
	// message type: Request (0)
	// Send route directly to `GameService.Hello` so gate will forward the Request to game nodes
	mbytes := encodeMessage(1, 0x00, "GameService.Hello", bb)
	dataPkg := encodePackage(4, mbytes)
	// debug: print first bytes of data package
	fmt.Printf("sending dataPkg len=%d head=% x\n", len(dataPkg), dataPkg[:16])
	if err := c.WriteMessage(websocket.BinaryMessage, dataPkg); err != nil {
		log.Fatal("write data:", err)
	}

	// read responses for a short time
	c.SetReadDeadline(time.Now().Add(5 * time.Second))
	for {
		_, m, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			return
		}
		// parse package header
		if len(m) < 4 {
			log.Println("recv too short")
			continue
		}
		pkgType := m[0]
		body := m[4:]
		if pkgType != 4 {
			fmt.Println("recv package type", pkgType)
			continue
		}
		// parse internal message
		mid, mtype, route, pdata := decodeMessage(body)
		fmt.Printf("recv message type=%d id=%d route=%s len=%d\n", mtype, mid, route, len(pdata))
		if mtype == 2 { // Response
			var resp struct {
				Message string `json:"message"`
			}
			if err := json.Unmarshal(pdata, &resp); err == nil {
				fmt.Println("Response message:", resp.Message)
			}
		}
	}
}

// decodeMessage decodes a nano Message from bytes and returns id,type,route,data
func decodeMessage(data []byte) (id uint64, mtype byte, route string, body []byte) {
	if len(data) < 1 {
		return 0, 0xff, "", nil
	}
	flag := data[0]
	mtype = byte((flag >> 1) & 0x07)
	offset := 1
	// id for Request/Response
	if mtype == 0 || mtype == 2 {
		var n uint64
		shift := uint(0)
		for {
			if offset >= len(data) {
				break
			}
			b := data[offset]
			offset++
			n |= uint64(b&0x7F) << shift
			if b < 128 {
				break
			}
			shift += 7
		}
		id = n
	}
	// route for Request/Notify/Push
	if mtype == 0 || mtype == 1 || mtype == 3 {
		if offset >= len(data) {
			return id, mtype, "", nil
		}
		rl := int(data[offset])
		offset++
		if offset+rl <= len(data) {
			route = string(data[offset : offset+rl])
			offset += rl
		}
	}
	if offset <= len(data) {
		body = data[offset:]
	}
	return
}

func encodePackage(t byte, body []byte) []byte {
	l := 4 + len(body)
	buf := make([]byte, l)
	buf[0] = t
	buf[1] = byte((len(body) >> 16) & 0xff)
	buf[2] = byte((len(body) >> 8) & 0xff)
	buf[3] = byte(len(body) & 0xff)
	copy(buf[4:], body)
	return buf
}

// encodeMessage encodes a nano internal Message (Request/Notify/Response/Push)
// following the server's message.Encode specification.
func encodeMessage(id uint64, msgType byte, route string, body []byte) []byte {
	// flag: type << 1
	flag := byte(msgType << 1)
	buf := make([]byte, 0, 2+len(route)+len(body)+8)
	buf = append(buf, flag)

	// if request or response, append varint id (little-endian 7-bit groups)
	if msgType == 0x00 || msgType == 0x02 { // Request or Response
		n := id
		for {
			b := byte(n % 128)
			n >>= 7
			if n != 0 {
				buf = append(buf, b+128)
			} else {
				buf = append(buf, b)
				break
			}
		}
	}

	// routable types: Request/Notify/Push include route
	if msgType == 0x00 || msgType == 0x01 || msgType == 0x03 {
		// route length (1 byte) + route bytes
		buf = append(buf, byte(len(route)))
		buf = append(buf, []byte(route)...)
	}

	// body
	if len(body) > 0 {
		buf = append(buf, body...)
	}
	return buf
}
