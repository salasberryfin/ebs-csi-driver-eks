package main

import (
	"crypto/sha1"
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
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

func (e *EBSCSIDriverEnableConfig) createOIDCProvider() {
	iamService, err := IAM()
	if err != nil {
		log.Fatalf("cannot start IAM session: %v", err)
	}
	output, err := iamService.ListOpenIDConnectProviders(&iam.ListOpenIDConnectProvidersInput{})
	if err != nil {
		log.Fatalf("listing OIDC providers: %v", err)
	}
	for _, prov := range output.OpenIDConnectProviderList {
		if strings.Contains(*prov.Arn, path.Base(*e.Cluster.Identity.Oidc.Issuer)) {
			log.Println("Found match:", *prov.Arn)
			e.OIDCConfig.OIDCProviderARN = *prov.Arn
			return
		}
	}
	log.Println("No matching providers found for:", path.Base(*e.Cluster.Identity.Oidc.Issuer))

	log.Println("Creating new OIDC provider")
	thumbprint, err := e.getIssuerThumbprint()
	if err != nil {
		log.Fatalf("getting issuer thumbprint: %v", err)
	}
	input := &iam.CreateOpenIDConnectProviderInput{
		ClientIDList:   []*string{aws.String(defaultAudience)},
		ThumbprintList: []*string{&thumbprint},
		Url:            e.Cluster.Identity.Oidc.Issuer,
		Tags:           []*iam.Tag{},
	}
	oidc, err := iamService.CreateOpenIDConnectProvider(input)
	if err != nil {
		log.Fatalf("creating OIDC provider: %v", err)
	}
	log.Println("created OIDC provider:", oidc)
	e.OIDCConfig.OIDCProviderARN = *oidc.OpenIDConnectProviderArn
}

func (e *EBSCSIDriverEnableConfig) getIssuerThumbprint() (string, error) {
	issuerURL, err := url.Parse(*e.Cluster.Identity.Oidc.Issuer)
	if err != nil {
		return "", fmt.Errorf("parsing issuer url: %w", err)
	}
	if issuerURL.Port() == "" {
		issuerURL.Host += ":443"
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
				MinVersion:         tls.VersionTLS12,
			},
			Proxy: http.ProxyFromEnvironment,
		},
	}
	resp, err := client.Get(issuerURL.String())
	if err != nil {
		return "", fmt.Errorf("querying oidc issuer endpoint %s: %w", issuerURL.String(), err)
	}
	defer resp.Body.Close()

	if resp.TLS == nil || len(resp.TLS.PeerCertificates) == 0 {
		return "", errors.New("unable to get OIDS issuers cert")
	}

	root := resp.TLS.PeerCertificates[len(resp.TLS.PeerCertificates)-1]
	return fmt.Sprintf("%x", sha1.Sum(root.Raw)), nil
}
