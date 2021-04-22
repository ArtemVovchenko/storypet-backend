package url

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParsePostgreConn(t *testing.T) {
	inp := "postgres://@localhost:5432/storypet"
	expected := "host=localhost port=5432 dbname=storypet"
	actual, err := ParsePostgreConn(inp)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}
