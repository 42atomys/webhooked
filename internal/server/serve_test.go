package server

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_NewServer(t *testing.T) {
	srv, err := NewServer(8080)
	assert.NoError(t, err)
	assert.NotNil(t, srv)

	srv, err = NewServer(0)
	assert.Error(t, err)
	assert.Nil(t, srv)
}

func Test_Serve(t *testing.T) {
	srv, err := NewServer(38081)
	assert.NoError(t, err)

	go func() {
		time.Sleep(2 * time.Second)
		assert.ErrorIs(t, srv.Shutdown(context.Background()), http.ErrServerClosed)
	}()

	assert.ErrorIs(t, srv.Serve(), http.ErrServerClosed)
}

func Test_validPort(t *testing.T) {
	assert := assert.New(t)

	var tests = []struct {
		input    int
		expected bool
	}{
		{8080, true},
		{1, true},
		{0, false},
		{-8080, false},
		{65535, false},
		{65536, false},
	}

	for _, test := range tests {
		assert.Equal(validPort(test.input), test.expected, "input: %d", test.input)
	}

}

func Test_newRouter(t *testing.T) {
	router := newRouter()
	assert.NotNil(t, router.NotFoundHandler)
}
