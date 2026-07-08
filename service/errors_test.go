package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidationError(t *testing.T) {
	err := ValidationError{Message: "bad input"}
	assert.Equal(t, "bad input", err.Error())
}

