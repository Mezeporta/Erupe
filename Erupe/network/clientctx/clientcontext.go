package clientctx

import "github.com/Solenataris/Erupe/common/stringsupport"

// ClientContext holds contextual data required for packet encoding/decoding.
type ClientContext struct {
	StrConv *stringsupport.StringConverter
}
