package cosem

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSecuritySetup(t *testing.T) {
	obis, _ := NewObisCodeFromString("0.0.43.0.0.255")
	clientSystemTitle := []byte("CLIENT")
	serverSystemTitle := []byte("SERVER")
	masterKey := []byte("master_key")
	guek := []byte("global_unicast_key")
	gak := []byte("global_auth_key")
	securitySetup, err := NewSecuritySetup(*obis, clientSystemTitle, serverSystemTitle, masterKey, guek, gak)
	assert.NoError(t, err)

	// Verify logical_name
	val, err := securitySetup.GetAttribute(1)
	assert.NoError(t, err)
	assert.Equal(t, *obis, val.(ObisCode))

	// Verify security_policy
	val, err = securitySetup.GetAttribute(2)
	assert.NoError(t, err)
	assert.Equal(t, PolicyNone, val.(SecurityPolicy))

	// Verify security_suite
	val, err = securitySetup.GetAttribute(3)
	assert.NoError(t, err)
	assert.Equal(t, SecuritySuite0, val.(SecuritySuite))

	// Verify client_system_title
	val, err = securitySetup.GetAttribute(4)
	assert.NoError(t, err)
	assert.Equal(t, clientSystemTitle, val.([]byte))

	// Verify server_system_title
	val, err = securitySetup.GetAttribute(5)
	assert.NoError(t, err)
	assert.Equal(t, serverSystemTitle, val.([]byte))

	// Verify keys
	assert.Equal(t, masterKey, securitySetup.MasterKey)
	assert.Equal(t, guek, securitySetup.GlobalUnicastKey)
	assert.Equal(t, gak, securitySetup.GlobalAuthenticationKey)
}

func TestSecurityPolicy(t *testing.T) {
	var sp SecurityPolicy
	sp = PolicyAuthenticatedRequest | PolicyEncryptedRequest
	assert.True(t, sp&PolicyAuthenticatedRequest != 0)
	assert.True(t, sp&PolicyEncryptedRequest != 0)
	assert.False(t, sp&PolicyDigitallySignedRequest != 0)

	sp = PolicyNone
	assert.False(t, sp&PolicyAuthenticatedRequest != 0)
}
