// +build ignore

package gorules

import "github.com/quasilyte/go-ruleguard/dsl"

func doubleClose(m dsl.Matcher) {
	m.Match(`$closer, $err := $_; $*_; $closer.Close(); $*_; $closer.Close()`).
		Where(m["closer"].Type.Implements("io.Closer")).
		Report("found double close for $closer")
}
