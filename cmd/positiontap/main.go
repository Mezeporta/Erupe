// Command positiontap is a TCP MITM proxy for the Monster Hunter Frontier
// channel-server MHF crypto stream. It terminates end-to-end crypto on both
// legs in memory, parses MSG_SYS_POSITION_OBJECT (0x0042) and the player
// state payload inside MSG_SYS_CAST_BINARY (0x0018, sub-type=0), and
// forwards every packet untouched to the upstream Erupe channel server.
//
// Usage:
//
//	cd server/Erupe
//	go build -o positiontap ./cmd/positiontap
//	./positiontap -listen 127.0.0.1:54001 \
//	              -upstream frontier.mogapedia.fr:54001 \
//	              -out positions.jsonl
//
// Then point the running mhf.exe (or your mhf-iel launcher) at listen instead
// of upstream. Requires the upstream Erupe server to be unchanged.
package main

import (
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"
	"time"

	"erupe-ce/common/byteframe"
	"erupe-ce/config"
	"erupe-ce/network"
	"go.uber.org/zap"
)

// opcodes we care about (network.MSG_*_BINARY/POSITION_OBJECT).
const (
	opcodeCastBinary     = network.MSG_SYS_CAST_BINARY     // 0x0018
	opcodePositionObject = network.MSG_SYS_POSITION_OBJECT // 0x0042
	opcodeCastedBinary   = network.MSG_SYS_CASTED_BINARY   // 0x001B (server->client mirror)
)

// Wire layout of every MHF channel packet (after the 14-byte CryptPacketHeader
// is decrypted off the wire):
//
//	[2 bytes opcode big-endian][payload...]
//
// See server/Erupe/network/mhfpacket/ and docs/network_protocol.md.

type loggedPos struct {
	Time   time.Time `json:"t"`
	Dir    string    `json:"dir"`               // "c2s" (client→server) or "s2c" (server→client)
	Source string    `json:"source"`            // "POSITION_OBJECT" or "PLAYER_STATE"
	ObjID  uint32    `json:"obj_id,omitempty"`  // for POSITION_OBJECT (server-side object id == charID)
	CharID uint32    `json:"char_id,omitempty"` // for PLAYER_STATE
	X      float32   `json:"x"`
	Y      float32   `json:"y"`
	Z      float32   `json:"z"`
}

func main() {
	listen := flag.String("listen", "127.0.0.1:54001", "local address the bot's mhf.exe connects to")
	upstream := flag.String("upstream", "frontier.mogapedia.fr:54001", "Erupe channel server upstream address")
	out := flag.String("out", "", "optional JSONL log of captured positions; omit for stderr-only")
	flag.Parse()

	// The bot is running version "ZZ" (G10-ZZ); anything >= F1 uses the
	// extended DataSize framing (F1+) so we hard-pin to config.ZZ. Different
	// versions would not interoperate against the live server anyway.
	mode := config.ZZ

	var sink io.Writer = os.Stdout
	if *out != "" {
		f, err := os.Create(*out)
		if err != nil {
			log.Fatalf("open -out: %v", err)
		}
		defer func() { _ = f.Close() }()
		sink = f
	}

	logger := zap.NewNop()

	ln, err := net.Listen("tcp", *listen)
	if err != nil {
		log.Fatalf("listen %s: %v", *listen, err)
	}
	log.Printf("positiontap: listening on %s, upstream %s, mode=%d", *listen, *upstream, int(mode))

	for {
		client, err := ln.Accept()
		if err != nil {
			log.Printf("accept: %v", err)
			continue
		}
		go handle(client, *upstream, mode, logger, sink)
	}
}

func handle(client net.Conn, upstream string, mode config.Mode, logger *zap.Logger, sink io.Writer) {
	defer func() { _ = client.Close() }()
	upstreamConn, err := net.Dial("tcp", upstream)
	if err != nil {
		log.Printf("dial upstream %s: %v", upstream, err)
		return
	}
	defer func() { _ = upstreamConn.Close() }()

	clientCC := network.NewCryptConn(client, mode, logger)
	serverCC := network.NewCryptConn(upstreamConn, mode, logger)

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		pipeEncrypted(clientCC, serverCC, "c2s", sink)
	}()
	go func() {
		defer wg.Done()
		pipeEncrypted(serverCC, clientCC, "s2c", sink)
	}()
	wg.Wait()
}

// pipeEncrypted reads packets from src, parses position-bearing ones into the
// JSONL sink, and re-encrypts them onto dst unmodified. Anywhere the parsing
// or forwarding fails we drop the connection: a partial stream would let an
// ack get lost and the client would soft-lock (per CLAUDE.md).
func pipeEncrypted(src, dst network.Conn, dir string, sink io.Writer) {
	for {
		plain, err := src.ReadPacket()
		if err != nil {
			if err != io.EOF {
				log.Printf("%s: read: %v", dir, err)
			}
			return
		}
		if len(plain) < 2 {
			log.Printf("%s: short packet (%d bytes)", dir, len(plain))
			return
		}
		opcode := network.PacketID(binary.BigEndian.Uint16(plain[:2]))
		payload := plain[2:]

		switch opcode {
		case opcodePositionObject:
			tryLogPositionObject(payload, dir, sink)
		case opcodeCastBinary, opcodeCastedBinary:
			tryLogCastBinary(payload, dir, sink)
		}

		if err := dst.SendPacket(plain); err != nil {
			log.Printf("%s: send: %v", dir, err)
			return
		}
	}
}

// MSG_SYS_POSITION_OBJECT payload (per mhfpacket/msg_sys_position_object.go):
//
//	[4 bytes obj_id BE][4 bytes x float32 BE][4 bytes y float32 BE][4 bytes z float32 BE]
func tryLogPositionObject(payload []byte, dir string, sink io.Writer) {
	if len(payload) < 16 {
		return
	}
	bf := byteframe.NewByteFrameFromBytes(payload)
	objID := bf.ReadUint32()
	x := bf.ReadFloat32()
	y := bf.ReadFloat32()
	z := bf.ReadFloat32()
	if bf.Err() != nil {
		return
	}
	writeLog(sink, loggedPos{
		Time: time.Now().UTC(), Dir: dir, Source: "POSITION_OBJECT",
		ObjID: objID, X: x, Y: y, Z: z,
	})
}

// MSG_SYS_CAST_BINARY payload (per mhfpacket/msg_sys_cast_binary.go):
//
//	[4 bytes unk][1 byte broadcast_type][1 byte message_type][2 bytes data_size][N bytes raw]
//
// PlayerStateBinary sub-packet (sub-type == 0) follows the 37-byte schema in
// docs/network_protocol.md:992-999 / client/OpenFrontier/scripts/network/
// packets/player_state_binary.gd:
//
//	[1 byte type=0][4 byte char_id][3*4 byte pos xyz][4 byte rot_y]
//	  [3*4 byte vel xyz][5 byte anim/flags/health] = 39 bytes total
//	  (the "37 bytes" in docs+OpenFrontier is off by 2; trust the field list
//	  and let a short read just skip the packet).
func tryLogCastBinary(payload []byte, dir string, sink io.Writer) {
	if len(payload) < 8 {
		return
	}
	bf := byteframe.NewByteFrameFromBytes(payload)
	bf.ReadUint32() // unk
	broadcastType := bf.ReadUint8()
	messageType := bf.ReadUint8()
	dataSize := bf.ReadUint16()
	if bf.Err() != nil {
		return
	}
	if messageType != 0 || broadcastType == 0xFF {
		// 0 == State / player position; 0xFF broadcast is admin-only.
		return
	}
	if dataSize < 5 || len(payload) < int(8+dataSize) {
		return
	}
	raw := payload[8 : 8+dataSize]
	if len(raw) < 1 || raw[0] != 0 {
		return
	}
	// raw[0] = sub-type (must be 0 == PlayerStateBinary)
	pbf := byteframe.NewByteFrameFromBytes(raw[1:])
	charID := pbf.ReadUint32()
	x := pbf.ReadFloat32()
	y := pbf.ReadFloat32()
	z := pbf.ReadFloat32()
	if pbf.Err() != nil {
		return
	}
	// We intentionally don't read rot_y/vel/animation here — the task is to
	// confirm position, not duplicate the OpenFrontier serializer.
	writeLog(sink, loggedPos{
		Time: time.Now().UTC(), Dir: dir, Source: "PLAYER_STATE",
		CharID: charID, X: x, Y: y, Z: z,
	})
}

func writeLog(sink io.Writer, r loggedPos) {
	b, err := json.Marshal(r)
	if err != nil {
		return
	}
	b = append(b, '\n')
	_, _ = sink.Write(b)
	_, _ = fmt.Fprintf(os.Stderr, "[%s] %s src=%-15s char=%d obj=%d xyz=(%.3f, %.3f, %.3f)\n",
		r.Dir, r.Time.Format("15:04:05.000"), r.Source, r.CharID, r.ObjID, r.X, r.Y, r.Z)
}
