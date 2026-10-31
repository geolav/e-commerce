package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	api := New(nil, nil, nil)

	assert.NotNil(t, api)
	assert.Nil(t, api.Loms)
	assert.Nil(t, api.Product)
	assert.Nil(t, api.Stocks)
}

func TestAPI_Struct(t *testing.T) {
	api := &API{}

	assert.NotNil(t, api)
}
