package main

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	eksv1 "github.com/rancher/eks-operator/pkg/eks"
	"github.com/rancher/eks-operator/pkg/eks/services"
)

func createIAMRole(oidc string) error {
	awsConfig := &aws.Config{}
	awsConfig.Region = aws.String("eu-central-1")
	accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	awsConfig.Credentials = credentials.NewStaticCredentials(accessKey, secretKey, "")
	sess, err := session.NewSession(awsConfig)
	if err != nil {
		fmt.Println("failed to create new AWS session")
		return err
	}

	displayName := "test-csalas-csi"
	cfService := services.NewCloudFormationService(sess)
	_, err = eksv1.CreateStack(eksv1.CreateStackOptions{
		CloudFormationService: cfService,
		StackName:             fmt.Sprintf("%s-ebs-csi-driver-role", displayName),
		DisplayName:           displayName,
		TemplateBody:          EBSCSIDriverTemplate,
		Capabilities:          []string{cloudformation.CapabilityCapabilityIam},
		Parameters:            []*cloudformation.Parameter{{ParameterKey: aws.String("OIDCProvider"), ParameterValue: aws.String(oidc)}},
	})
	if err != nil {
		fmt.Println("Error:", err)
		// If there was an error creating the driver role stack, return an empty role arn and the error
		return err
	}

	return nil
}
