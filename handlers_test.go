package gnosis

import (
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStripRequestRouting(t *testing.T) {
	r, err := http.NewRequest("GET", "/some/weird/path.gif", *new(io.Reader))
	assert.NoError(t, err)

	r, err = http.NewRequest("GET", "/some/weird/path.gif", *new(io.Reader))
	request, err := stripRequestRouting("/some/", r)
	assert.NoError(t, err)
	if err == nil {
		assert.Equal(t, "/weird/path.gif", request.URL.Path)
	}

	r, err = http.NewRequest("GET", "/some/weird/path.gif", *new(io.Reader))
	request, err = stripRequestRouting("/some/weird/", r)
	assert.NoError(t, err)
	if err == nil {
		assert.Equal(t, "/path.gif", request.URL.Path)
	}

	r, err = http.NewRequest("GET", "/some/weird/path.gif", *new(io.Reader))
	request, err = stripRequestRouting("some/weird/", r)
	assert.EqualError(t, err, "passed a request route that does not start in a /")

	r, err = http.NewRequest("GET", "/some/weird/path.gif", *new(io.Reader))
	request, err = stripRequestRouting("/some/weird", r)
	assert.EqualError(t, err, "passed a request route that does not end in a /")

	r, err = http.NewRequest("GET", "/some/weird/path.gif", *new(io.Reader))
	request, err = stripRequestRouting("/some/weird/path_is_too_long_by_quite_a_bit/", r)
	assert.EqualError(t, err, "request routing path longer than request path")

	r, err = http.NewRequest("GET", "/some/weird/path.gif", *new(io.Reader))
	request, err = stripRequestRouting("/other/junk/", r)
	assert.EqualError(t, err, "request does not match up to the routed path")
}
