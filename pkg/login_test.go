package pkg

import (
	"testing"
)

func Test_VaultAwsLogin(t *testing.T) {
	token, err := VaultAWSLogin("http://127.0.0.1:8200", "default", "eks-staging-cluster", "/home/acohea/.aws/credentials")

	if err != nil {
		t.Error(err)
	}
	if token == "" {
		t.Errorf("expected a non-empty token")
	}

	t.Errorf("%s", token)
}
