package pkg

import (
	"errors"
	"fmt"

	"github.com/hashicorp/vault/api"
)

func GetVaultClient(url, token string) (*api.Client, error) {
	config := api.DefaultConfig()
	config.Address = url
	client, err := api.NewClient(config)
	if err != nil {
		return nil, err
	}
	client.SetToken(token)

	return client, nil
}

func GetCassandraCreds(client *api.Client, cluster, dbRole string) (string, string, error) {
	creds, err := client.Logical().Read(fmt.Sprintf("%s/creds/%s", cluster, dbRole))
	if err != nil {
		return "", "", err
	}

	usernameValue, ok := creds.Data["username"]
	if !ok {
		return "", "", errors.New("username not returned from Vault")
	}

	username, ok := usernameValue.(string)
	if !ok {
		return "", "", errors.New("username is not a string")
	}

	passwordValue, ok := creds.Data["password"]
	if !ok {
		return "", "", errors.New("password not returned from Vault")
	}

	password, ok := passwordValue.(string)
	if !ok {
		return "", "", errors.New("password is not a string")
	}

	return username, password, nil
}
