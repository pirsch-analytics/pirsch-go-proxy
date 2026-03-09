package proxy

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAcceptRequest(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "https://proxy.com/hit?url=https://example.com/foo/bar&code=asdf1234", nil)
	assert.True(t, acceptRequest(client{
		filter: []FilterFunc{},
	}, req))
	assert.False(t, acceptRequest(client{
		filter: []FilterFunc{
			NewHostnameFilter([]string{"test.com"}),
		},
	}, req))
	assert.False(t, acceptRequest(client{
		filter: []FilterFunc{
			NewPathFilter([]string{"/some/path"}),
		},
	}, req))
	assert.False(t, acceptRequest(client{
		filter: []FilterFunc{
			NewIdentificationCodeFilter([]string{"1234asdf"}),
		},
	}, req))
	assert.True(t, acceptRequest(client{
		filter: []FilterFunc{
			NewHostnameFilter([]string{"www.example.com", "example.com"}),
		},
	}, req))
	assert.True(t, acceptRequest(client{
		filter: []FilterFunc{
			NewPathFilter([]string{"/some/path", "/foo/bar"}),
		},
	}, req))
	assert.True(t, acceptRequest(client{
		filter: []FilterFunc{
			NewIdentificationCodeFilter([]string{"1234asdf", "asdf1234"}),
		},
	}, req))
	assert.True(t, acceptRequest(client{
		filter: []FilterFunc{
			NewHostnameFilter([]string{"www.example.com", "example.com"}),
			NewPathFilter([]string{"/some/path", "/foo/bar"}),
			NewIdentificationCodeFilter([]string{"1234asdf", "asdf1234"}),
		},
	}, req))
	assert.False(t, acceptRequest(client{
		filter: []FilterFunc{
			NewHostnameFilter([]string{"www.example.com", "example.com"}),
			NewPathFilter([]string{"/some/path"}),
			NewIdentificationCodeFilter([]string{"1234asdf", "asdf1234"}),
		},
	}, req))
}
