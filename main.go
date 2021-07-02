package main

import (
	"log"
	"os"

	"github.com/AdrienneCohea/vault-cqlsh/pkg"
	"github.com/manifoldco/promptui"
	home "github.com/mitchellh/go-homedir"
)

func main() {
	connections, err := pkg.LoadConnections()
	if err != nil {
		log.Fatal(err)
	}

	hostPrompt := promptui.Select{
		Label: "Cassandra Instance",
		Items: pkg.GetDescriptions(connections),
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

	token, err := pkg.VaultAWSLogin(c.AuthVaultUrl, c.AuthAwsProfile, c.AuthAwsRole, c.AuthVaultHeader, awsCredentialsFile)
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
