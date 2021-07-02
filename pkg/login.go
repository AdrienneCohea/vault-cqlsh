package pkg

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/credentials"
	v4 "github.com/aws/aws-sdk-go/aws/signer/v4"
	"github.com/hashicorp/vault/api"
)

type AwsLoginRequest struct {
	Role                   string `json:"role"`
	Identity               string `json:"identity"`
	IAMHttpRequestMethod   string `json:"iam_http_request_method"`
	IAMRequestUrl          string `json:"iam_request_url"`
	IAMRequestBody         string `json:"iam_request_body"`
	IAMRequestHeaders      string `json:"iam_request_headers"`
	IAMServerIdHeaderValue string `json:"iam_server_id_header_value"`
}

type LoginResponse struct {
	Auth struct {
		Renewable     bool `json:"renewable"`
		LeaseDuration int  `json:"lease_duration"`
		Metadata      struct {
			RoleTagMaxTTL string `json:"role_tag_max_ttl"`
			InstanceID    string `json:"instance_id"`
			AmiID         string `json:"ami_id"`
			Role          string `json:"role"`
			AuthType      string `json:"auth_type"`
		} `json:"metadata"`
		Policies    []string `json:"policies"`
		Accessor    string   `json:"accessor"`
		ClientToken string   `json:"client_token"`
	} `json:"auth"`
}

func VaultAWSLogin(vaultUrl, profile, role, vaultHeaderValue, sharedCredentialsFile string) (string, error) {
	login, err := newVaultAwsLoginRequest(sharedCredentialsFile, profile, role, vaultHeaderValue)
	if err != nil {
		return "", err
	}

	return login.auth(vaultUrl)
}

func newVaultAwsLoginRequest(filename, profile, role, vaultHeaderValue string) (*AwsLoginRequest, error) {
	headers, url, body, err := createGetCallerIdentity(filename, profile, vaultHeaderValue)
	if err != nil {
		return nil, err
	}

	login := &AwsLoginRequest{
		IAMHttpRequestMethod: "POST",
		IAMRequestHeaders:    headers,
		IAMRequestUrl:        url,
		IAMRequestBody:       body,
		Role:                 role,
	}

	return login, nil
}

func (login *AwsLoginRequest) auth(vaultUrl string) (string, error) {
	loginRequest, err := json.Marshal(login)
	if err != nil {
		return "", err
	}

	u, err := url.Parse(vaultUrl)
	if err != nil {
		return "", err
	}
	u.Path = "/v1/auth/aws/login"

	req, err := http.NewRequest("POST", u.String(), bytes.NewBuffer(loginRequest))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	switch resp.StatusCode {
	case 200:
		log.Println("Vault: AWS login succeeded")
		loginResponse := LoginResponse{}
		err = json.Unmarshal(body, &loginResponse)
		if err != nil {
			return "", fmt.Errorf("%v in body %s", err, body)
		}
		return loginResponse.Auth.ClientToken, nil
	default:
		log.Println("Vault: AWS login error")
		loginError := api.ResponseError{}
		err = json.Unmarshal(body, &loginError)
		if err != nil {
			return "", fmt.Errorf("%v in body %s", err, body)
		}
		return "", errors.New(loginError.Error())
	}
}

func createGetCallerIdentity(filename, profile, vaultHeaderValue string) (string, string, string, error) {
	creds := credentials.NewSharedCredentials(filename, profile)
	signer := v4.NewSigner(creds)

	//TODO: Fix this mess by using the AWS SDK to make the GetCallerIdentity request instead.

	region := "us-east-1"
	host := fmt.Sprintf("sts.%s.amazonaws.com", region)
	stsUrl := fmt.Sprintf("https://%s", host)
	body := "Action=GetCallerIdentity&Version=2011-06-15"
	reader := strings.NewReader(body)

	callerIdentity, _ := http.NewRequest("POST", stsUrl, reader)
	callerIdentity.Header.Set("Accept-Encoding", "identity")
	callerIdentity.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	callerIdentity.Header.Set("User-Agent", "vault-cqlsh")
	callerIdentity.Header.Add("X-Vault-AWS-IAM-Server-ID", vaultHeaderValue)
	_, err := signer.Sign(callerIdentity, reader, "sts", region, time.Now().UTC())
	if err != nil {
		return "", "", "", err
	}
	jsonHeaders, err := json.Marshal(callerIdentity.Header)
	if err != nil {
		return "", "", "", err
	}

	//We are not actually calling GetCallerIdentity. Vault is doing that, and
	//we simply return the relevant headers, URN, and body for the Vault AWS login.
	iamRequestheaders := base64.StdEncoding.EncodeToString(jsonHeaders)
	iamRequestUrl := base64.StdEncoding.EncodeToString([]byte(stsUrl))
	iamRequestBody := base64.StdEncoding.EncodeToString([]byte(body))

	return iamRequestheaders, iamRequestUrl, iamRequestBody, nil
}
