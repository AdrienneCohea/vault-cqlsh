package main

import (
	"log"
	"os"

	"github.com/AdrienneCohea/vault-cqlsh/pkg"
	"github.com/manifoldco/promptui"
	home "github.com/mitchellh/go-homedir"
)

func main() {
	connections := []CassandraConnectionInfo{
		{
			CassandraHost:  "172.17.0.2",
			AuthVaultUrl:   "http://127.0.0.1:8200",
			AuthAwsProfile: "default",
			AuthAwsRole:    "eks-staging-cluster",
			VaultMount:     "testcluster",
		},
	}
	hostPrompt := promptui.Select{
		Label: "Cassandra Instance",
		Items: getHosts(connections),
	}

	index, _, err := hostPrompt.Run()
	if err != nil {
		log.Fatal(err)
	}

	accessLevelPrompt := promptui.Select{
		Label: "Access Level",
		Items: []string{"readonly", "readwrite"},
	}

	_, accessLevel, err := accessLevelPrompt.Run()
	if err != nil {
		log.Fatal(err)
	}

	c := connections[index]

	awsCredentialsFile, err := home.Expand("~/.aws/credentials")
	if err != nil {
		log.Fatal(err)
	}

	token, err := pkg.VaultAWSLogin(c.AuthVaultUrl, c.AuthAwsProfile, c.AuthAwsRole, awsCredentialsFile)
	if err != nil {
		log.Fatal(err)
	}

	client, err := pkg.GetVaultClient(c.AuthVaultUrl, token)
	if err != nil {
		log.Fatal(err)
	}

	username, password, err := pkg.GetCassandraCreds(client, c.VaultMount, accessLevel)
	if err != nil {
		log.Fatal(err)
	}

	pkg.ExecCqlsh(username, password, c.CassandraHost, "9042", os.Args[1:])
}

type CassandraConnectionInfo struct {
	CassandraHost  string `json:"host"`
	AuthVaultUrl   string `json:"auth_vault_url"`
	AuthAwsProfile string `json:"auth_aws_profile"`
	AuthAwsRole    string `json:"auth_aws_role"`
	VaultMount     string `json:"vault_mount"`
}

func getHosts(connections []CassandraConnectionInfo) []string {
	hosts := make([]string, len(connections))

	i := 0

	for _, k := range connections {
		hosts[i] = k.CassandraHost
		i++
	}

	return hosts
}
