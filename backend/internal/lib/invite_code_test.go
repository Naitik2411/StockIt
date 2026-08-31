package lib_test

import (
	"testing"

	"github.com/Naitik2411/stockit/internal/lib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateInviteCode_LengthAndCharset(t *testing.T) {
	code, err := lib.GenerateInviteCode()
	require.NoError(t, err)

	// 5 random bytes → base32 without padding ≈ 8 chars
	assert.GreaterOrEqual(t, len(code), 8)
	assert.LessOrEqual(t, len(code), 10)

	for _, r := range code {
		ok := (r >= 'A' && r <= 'Z') || (r >= '2' && r <= '7')
		assert.Truef(t, ok, "unexpected char %q in %q", r, code)
	}
}

func TestGenerateInviteCode_Unique(t *testing.T) {
	seen := make(map[string]struct{}, 100)
	for i := 0; i < 100; i++ {
		code, err := lib.GenerateInviteCode()
		require.NoError(t, err)
		_, exists := seen[code]
		assert.False(t, exists, "duplicate code %s", code)
		seen[code] = struct{}{}
	}
}
