package solver

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetValue(t *testing.T) {
	tests := map[string]struct {
		value       int
		expectedErr error
	}{
		"Good Value": {
			value: 5,
		},
		"Bad Value": {
			value:       -1,
			expectedErr: errInvalidValue,
		},
	}

	for name, test := range tests {
		t.Run(name, func(tt *testing.T) {
			c := NewCell()
			err := c.SetValue(test.value)

			if test.expectedErr == nil {
				assert.NoError(tt, err)
			} else {
				assert.ErrorIs(tt, err, test.expectedErr)
			}
		})
	}
}
