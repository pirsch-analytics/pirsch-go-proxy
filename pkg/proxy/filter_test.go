package proxy

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHostnameFilter(t *testing.T) {
	filter := NewHostnameFilter([]string{
		"filtered.com",
		"regex:[a-z]+\\.filtered\\.com",
	})
	u, _ := url.Parse("https://proxy.com/hit?url=https://example.com/")
	assert.False(t, filter(u))
	u, _ = url.Parse("https://proxy.com/hit?url=https://example.com/blog/article")
	assert.False(t, filter(u))
	u, _ = url.Parse("https://proxy.com/hit?url=https://filtered.com/blog/filtered")
	assert.True(t, filter(u))
	u, _ = url.Parse("https://proxy.com/hit?url=https://sub.filtered.com")
	assert.True(t, filter(u))
	u, _ = url.Parse("https://proxy.com/hit?url=https://01.filtered.com")
	assert.False(t, filter(u))
}

func TestPathFilter(t *testing.T) {
	filter := NewPathFilter([]string{
		"/blog/filtered",
		"regex:\\/glossary\\/[0-9]+",
	})
	u, _ := url.Parse("https://proxy.com/hit?url=https://example.com/")
	assert.False(t, filter(u))
	u, _ = url.Parse("https://proxy.com/hit?url=https://example.com/blog/article")
	assert.False(t, filter(u))
	u, _ = url.Parse("https://proxy.com/hit?url=https://example.com/blog/filtered")
	assert.True(t, filter(u))
	u, _ = url.Parse("https://proxy.com/hit?url=https://example.com/glossary/9342589")
	assert.True(t, filter(u))
	u, _ = url.Parse("https://proxy.com/hit?url=https://example.com/glossary/e9342589")
	assert.False(t, filter(u))
}

func TestIdentificationCodeFilter(t *testing.T) {
	filter := NewIdentificationCodeFilter([]string{
		"abc123",
		"efg456",
	})
	u, _ := url.Parse("https://proxy.com/hit?code=123456")
	assert.False(t, filter(u))
	u, _ = url.Parse("https://proxy.com/hit?code=abc123")
	assert.True(t, filter(u))
	u, _ = url.Parse("https://proxy.com/hit?code=efg456")
	assert.True(t, filter(u))
}
