package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
)

type Credentials struct {
	AccessKey string
	SecretKey string
	Region    string
}

var creds Credentials

func createSession(accessKey, secretKey, region string) (*session.Session, error) {
	cfg := aws.NewConfig()
	cfg.Credentials = credentials.NewStaticCredentials(accessKey, secretKey, "")
	cfg.Region = aws.String(region)

	cfg.CredentialsChainVerboseErrors = aws.Bool(true)
	sess, err := session.NewSessionWithOptions(session.Options{
		Config:            *cfg,
		SharedConfigState: session.SharedConfigDisable,
	})
	if err != nil {
		return nil, fmt.Errorf("creating AWS session: %w", err)
	}
	sess.Handlers.Build.PushFront(request.WithAppendUserAgent("eks-ebs-enable"))

	return sess, nil
}

func init() {
	if access_key, exist := os.LookupEnv("AWS_ACCESS_KEY_ID"); !exist {
		log.Fatal("AWS_ACCESS_KEY_ID not set")
	} else {
		creds.AccessKey = access_key
	}
	if secret_key, exist := os.LookupEnv("AWS_SECRET_ACCESS_KEY"); !exist {
		log.Fatal("AWS_SECRET_ACCESS_KEY not set")
	} else {
		creds.SecretKey = secret_key
	}
	if region, exist := os.LookupEnv("AWS_REGION"); !exist {
		log.Fatal("AWS_REGION not set")
	} else {
		creds.Region = region
	}
}

func main() {
	/*
		1. OIDC provider
		2. IAM role
			a. Trust relationship with OIDC
			b. Permissions policy for volume actions
		3. Install EKS addon / Deploy driver
		4. Annotate Service Account `ebs-csi-controller-sa` with IAM role name
		5. Restart `ebs-csi-controller` deployment
	*/
	getEKSCluster()
	log.Printf("Creating OIDC provider\n")
	//createOIDCProvider()
}
