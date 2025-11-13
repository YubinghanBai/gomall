package password

import (
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"testing"
)

func TestPassword(t *testing.T) {
	password := "test123456"

	hashPassword1, err := HashPassword(password)
	require.NoError(t, err)
	require.NotEmpty(t, hashPassword1)

	err = VerifyPassword(hashPassword1, password)
	require.NoError(t, err)

	wrongPassword := "wrong123456"
	err = VerifyPassword(hashPassword1, wrongPassword)
	require.EqualError(t, err, bcrypt.ErrMismatchedHashAndPassword.Error())

	hashPassword2, err := HashPassword(password)
	require.NoError(t, err)
	require.NotEmpty(t, hashPassword2)
	require.NotEqual(t, hashPassword1, hashPassword2)

}
