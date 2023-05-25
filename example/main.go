package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/rancher/eks-operator/pkg/eks/services"
)

type Credentials struct {
	AccessKey string
	SecretKey string
	Region    string
}

type AWSConfig struct {
	CFService   services.CloudFormationServiceInterface
	EKSSservice services.EKSServiceInterface
	IAMService  services.IAMServiceInterface
}

type ClusterSpec struct {
	Name         string
	EBSCSIDriver *bool
}

type ClusterConfig struct {
	Spec ClusterSpec
	AWS  AWSConfig
}

type OIDCProviderStep struct {
	OIDCProviderARN string
}

type EBSCSIDriverEnableConfig struct {
	Cluster    *eks.Cluster
	Spec       ClusterSpec
	OIDCConfig OIDCProviderStep
}

var clusterConfig ClusterConfig

func init() {
	var accessKey, secretKey, region string
	var exist bool
	if accessKey, exist = os.LookupEnv("AWS_ACCESS_KEY_ID"); !exist {
		log.Fatal("AWS_ACCESS_KEY_ID not set")
	}
	if secretKey, exist = os.LookupEnv("AWS_SECRET_ACCESS_KEY"); !exist {
		log.Fatal("AWS_SECRET_ACCESS_KEY not set")
	}
	if region, exist = os.LookupEnv("AWS_REGION"); !exist {
		log.Fatal("AWS_REGION not set")
	}
	awsConfig := &aws.Config{}
	awsConfig.Region = aws.String(region)
	awsConfig.Credentials = credentials.NewStaticCredentials(accessKey, secretKey, "")
	sess, err := session.NewSession(awsConfig)
	if err != nil {
		fmt.Println("failed to create new AWS session")
		os.Exit(1)
	}

	cfService := services.NewCloudFormationService(sess)
	eksServices := services.NewEKSService(sess)
	iamServices := services.NewIAMService(sess)
	clusterConfig.AWS = AWSConfig{
		CFService:   cfService,
		EKSSservice: eksServices,
		IAMService:  iamServices,
	}
}

func main() {
	/*
		- [x] OIDC provider
		- [x] IAM role
			a. Trust relationship with OIDC
			b. Permissions policy for volume actions
		- [ ] Install EKS addon / Deploy driver
		- [ ] Annotate Service Account `ebs-csi-controller-sa` with IAM role name
		- [ ] Restart `ebs-csi-controller` deployment
	*/
	var clusterName string
	var exist bool
	if clusterName, exist = os.LookupEnv("EKS_CLUSTER_NAME"); !exist {
		log.Fatalf("EKS_CLUSTER_NAME not set")
		clusterName = "test-eks-cluster"
	}
	clusterSpec := ClusterSpec{
		Name:         clusterName,
		EBSCSIDriver: aws.Bool(true),
	}
	clusterConfig.Spec = clusterSpec
	eksCluster := newEKSCluster()
	config := EBSCSIDriverEnableConfig{
		Cluster: eksCluster,
		Spec:    clusterSpec,
	}
	log.Printf("Looking for existing OIDC providers\n")
	oidc := config.createOIDCProvider()
	log.Println("The ID of the OIDC provider is:", oidc)
	roleArn, err := createIAMRole(oidc)
	if err != nil {
		log.Fatalf("creating iam role: %v", err)
	}
	addonArn, err := checkEBSAddon()
	if err != nil {
		log.Fatalf("checking if ebs addon is installed: %v", err)
	}
	if addonArn == "" {
		// addon not installed
		log.Println("EBS CSI driver addon not installed, installing now")
		err = installEBSAddon(roleArn)
	}
	addonArn, err = checkEBSAddon()
	if err != nil {
		log.Fatalf("checking if ebs addon is installed: %v", err)
	}
	if addonArn != "" {
		log.Println("EBS CSI driver addon installed")
	}
}
