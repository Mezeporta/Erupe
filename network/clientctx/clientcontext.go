package clientctx

import "erupe-ce/config"

// ClientContext holds contextual data required for packet encoding/decoding.
type ClientContext struct {
	RealClientMode _config.Mode
}
