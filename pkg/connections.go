package pkg

import (
	"encoding/json"
	"io/ioutil"
)

func LoadConnections() ([]CassandraConnectionInfo, error) {
	contents, err := ioutil.ReadFile("connections.json")
	if err != nil {
		return nil, err
	}

	configurations := &CassandraConfigurations{}
	err = json.Unmarshal(contents, &configurations)
	if err != nil {
		return nil, err
	}

	return configurations.Clusters, nil
}

type CassandraConfigurations struct {
	Clusters []CassandraConnectionInfo `json:"clusters"`
}

type CassandraConnectionInfo struct {
	Description     string `json:"description"`
	CassandraHost   string `json:"cassandra_host"`
	AuthVaultUrl    string `json:"auth_vault_url"`
	AuthVaultHeader string `json:"auth_vault_header"`
	AuthAwsProfile  string `json:"auth_aws_profile"`
	AuthAwsRole     string `json:"auth_aws_role"`
	VaultMount      string `json:"vault_mount"`
}

func GetDescriptions(connections []CassandraConnectionInfo) []string {
	descriptions := make([]string, len(connections))

	i := 0

	for _, conn := range connections {
		descriptions[i] = conn.Description
		i++
	}

	return descriptions
}
