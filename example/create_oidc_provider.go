package main

import (
	"log"

	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
)

func IAM() (iamiface.IAMAPI, error) {
	session, err := createSession(creds.AccessKey, creds.SecretKey, creds.Region)
	if err != nil {
		return nil, err
	}

	return iam.New(session), nil
}

func createOIDCProvider() {
	iamService, err := IAM()
	if err != nil {
		log.Fatalf("cannot start IAM session: %v", err)
	}
	output, err := iamService.ListOpenIDConnectProviders(&iam.ListOpenIDConnectProvidersInput{})
	if err != nil {
		log.Fatalf("listing OIDC providers: %v", err)
	}
	for _, prov := range output.OpenIDConnectProviderList {
		log.Println(prov)
	}
}
