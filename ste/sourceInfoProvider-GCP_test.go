package ste

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Azure/azure-storage-azcopy/v10/common"
)

// setTestGCPJSONKey points the package-level jsonKey at a throwaway service account key so URLs can be signed offline.
func setTestGCPJSONKey(t *testing.T) {
	a := assert.New(t)
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	a.NoError(err)
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	keyJSON, err := json.Marshal(map[string]string{
		"type":         "service_account",
		"client_email": "test@example.iam.gserviceaccount.com",
		"private_key":  string(keyPEM),
		"token_uri":    "https://oauth2.googleapis.com/token",
	})
	a.NoError(err)

	prev := jsonKey
	jsonKey = keyJSON
	t.Cleanup(func() { jsonKey = prev })
}

func TestGCPPreSignedSourceURLUserProject(t *testing.T) {
	setTestGCPJSONKey(t)
	p := &gcpSourceInfoProvider{gcpURLParts: common.GCPURLParts{BucketName: "bucket", ObjectKey: "dir/object.txt"}}

	t.Run("unset", func(t *testing.T) {
		a := assert.New(t)
		t.Setenv(common.EEnvironmentVariable.GCPUserProject().Name, "")
		signed, err := p.PreSignedSourceURL()
		a.NoError(err)
		u, err := url.Parse(signed)
		a.NoError(err)
		a.False(u.Query().Has("userProject"))
	})

	t.Run("set", func(t *testing.T) {
		a := assert.New(t)
		t.Setenv(common.EEnvironmentVariable.GCPUserProject().Name, "billing-project")
		signed, err := p.PreSignedSourceURL()
		a.NoError(err)
		u, err := url.Parse(signed)
		a.NoError(err)
		a.Equal("billing-project", u.Query().Get("userProject"))
		a.NotEmpty(u.Query().Get("X-Goog-Signature"))
	})
}
