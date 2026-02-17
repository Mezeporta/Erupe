package clientctx

import (
	"testing"
)

// TestClientContext_Exists verifies that the ClientContext type exists
// and can be instantiated.
func TestClientContext_Exists(t *testing.T) {
	ctx := ClientContext{}
	_ = ctx
}
