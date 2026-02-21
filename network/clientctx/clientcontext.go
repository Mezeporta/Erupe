package clientctx

import cfg "erupe-ce/config"

// ClientContext holds contextual data required for packet encoding/decoding.
type ClientContext struct {
	RealClientMode cfg.Mode
}
