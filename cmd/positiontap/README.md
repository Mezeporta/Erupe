# positiontap — Erupe-side TCP MITM for position packets

`positiontap` is a small Go TCP server that sits between a real `mhf.exe`
client and an Erupe channel server. For every connection it accepts on
`-listen`, it dials `-upstream`, then runs two goroutines that decrypt
each direction independently and re-encrypt for the other. Anywhere in
either pipeline it sees a position-bearing packet, it logs the parsed
`x / y / z` (and where applicable the `obj_id` / `char_id`) to a JSONL
file plus a one-line stderr summary.

It is _not_ an Erupe feature — it does not modify any packet. It is purely
a passive observer that lives in the network path. Pair it with
`tools/position-tap/snap_match_struct.py` (or any of the other readers
in `tools/position-tap/`) and you can verify the position intent from two
independent sources (wire bytes vs. process memory of the very same
running `mhfo-hd.dll`).

## Build

```
cd server/Erupe
go build -mod=mod -o positiontap ./cmd/positiontap/
```

## Run

```
./positiontap \
    -listen   127.0.0.1:54001  \
    -upstream frontier.mogapedia.fr:54001 \
    -out      /tmp/positions.jsonl
```

The `-listen` and `-upstream` arguments are precise: the local listener
should be the address the real client (`mhf.exe`, routed via
`iptables -t nat -A OUTPUT -p tcp -d <real upstream> --dport 54001 -j DNAT --to-destination <listen>`)
will end up at; the `-upstream` argument is the real Erupe channel
server.

The `-out` flag is the JSONL destination (one record per parsed position
packet). When omitted, positiontap goes silent on stdout and writes only
human-readable summaries to stderr.

## Packets parsed

| Opcode | Source | Fields captured | Why |
|---|---|---|---|
| `0x0042` `MSG_SYS_POSITION_OBJECT` | both directions | `obj_id, x, y, z` (16-byte payload) | Object-position broadcast (traps, barrels, NPC, player-self once the server relays). |
| `0x0018` `MSG_SYS_CAST_BINARY` | client → server | when `MessageType == 0` (player state): `char_id, x, y, z` from `PlayerStateBinary` sub-payload | The bot's own periodic state updates. |
| `0x001B` `MSG_SYS_CASTED_BINARY` | server → client | same sub-payload parsing | Server rebroadcasting other characters' positions to this client. |

`PlayerStateBinary` is the 37- (or 39-) byte sub-packet inside `MSG_SYS_CAST_BINARY`. Its layout (per `docs/network_protocol.md:992-999` and
`client/OpenFrontier/scripts/network/packets/player_state_binary.gd`):

```
[ 1 byte type=0 ]
[ 4 byte char_id ]      reader reads char_id
[ 3 × 4 byte pos ]      reader reads x, y, z
[ 1 × 4 byte rot_y ]
[ 3 × 4 byte velocity ]
[ 5 byte anim/flags/health ]
```

positiontap parses and discards everything after `z` — its goal is to
prove the position stream, not to be a complete protocol observer.

## iptables setup

`positiontap`'s outbound dial to upstream is itself subject to the same
DNAT rule that brought the bot's traffic to the proxy. To exempt
positiontap's outbound (so it can reach upstream, not loop back to
itself), positiontap must run as root, and the rule chain must include
a return rule before the DNAT:

```bash
sudo iptables -t nat -A OUTPUT -p tcp -d <upstream_host> --dport 54001 \
    -j DNAT --to-destination 127.0.0.1:54001
sudo iptables -t nat -I OUTPUT 1 -p tcp --dport 54001 \
    -m owner --uid-owner root -j RETURN
```

The `tools/position-tap/pos_runner.c` setuid wrapper (compiled and
installed as `/tmp/run-positiontap`, owner root:root, mode 4755) makes
either of:

- sudo-launched `positiontap`
- setuid-wrapped invocation

work for the exemption — both run as uid 0.

## What it does NOT do

- It does not modify packets. It only forwards them, byte-for-byte, after
  decrypting on one side and re-encrypting on the other. Each side's
  `network.CryptConn` keeps its own key state and increments as the real
  game does, so the upstream Erupe and the bot's mhf.exe see the same
  cryptographic stream a direct connection would have produced.
- It does not sign or authenticate. It plays no role in the protocol's
  handshake (sign / DSGN / entrance) — only the channel-server MHF
  crypto tunnel.
- It does not capture *all* state — only position-relevant packets. If
  you need chat, quest, mail-notify, etc., extend `tryLog*` the same way
  `tryLogCastBinary` is structured.

## Reading the JSONL

Each row is a single decoded position packet:

```json
{
  "t": "2026-07-11T17:59:54.037Z",
  "dir": "c2s",
  "source": "POSITION_OBJECT",
  "obj_id": 65537,
  "x": 10313.229, "y": 201.057, "z": 7388.005
}
```

`dir` is `c2s` (client → server) or `s2c` (server → client). `source`
distinguishes the wire opcode (POSITION_OBJECT vs PLAYER_STATE). Use
`(dir, obj_id or char_id)` as a stable identity key for correlating the
same entity across packets.
