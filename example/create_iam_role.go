package main

import (
	"bytes"
	"fmt"
	"html/template"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	eksv1 "github.com/rancher/eks-operator/pkg/eks"
	"github.com/rancher/eks-operator/pkg/eks/services"
)

func createIAMRole(oidc string) (string, error) {
	awsConfig := &aws.Config{}
	awsConfig.Region = aws.String("eu-central-1")
	accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	awsConfig.Credentials = credentials.NewStaticCredentials(accessKey, secretKey, "")
	sess, err := session.NewSession(awsConfig)
	if err != nil {
		fmt.Println("failed to create new AWS session")
		return "", err
	}

	cfService := services.NewCloudFormationService(sess)
	templateData := struct {
		Region     string
		ProviderID string
	}{
		Region:     *awsConfig.Region,
		ProviderID: oidc,
	}
	tmpl, err := template.New("ebsrole").Parse(EBSCSIDriverTemplate)
	if err != nil {
		return "", fmt.Errorf("parsing ebs role template: %w", err)
	}
	buf := &bytes.Buffer{}
	if execErr := tmpl.Execute(buf, templateData); execErr != nil {
		return "", fmt.Errorf("executing ebs role template: %w", err)
	}
	finalTemplate := buf.String()
	displayName := "test-csalas-csi"

	output, err := eksv1.CreateStack(eksv1.CreateStackOptions{
		CloudFormationService: cfService,
		StackName:             fmt.Sprintf("%s-ebs-csi-driver-role", displayName),
		DisplayName:           displayName,
		TemplateBody:          finalTemplate,
		Capabilities:          []string{cloudformation.CapabilityCapabilityIam},
		Parameters:            []*cloudformation.Parameter{},
	})
	if err != nil {
		fmt.Println("Error:", err)
		// If there was an error creating the driver role stack, return an empty role arn and the error
		return "", err
	}
	roleARN := getParameterValueFromOutput("EBSCSIDriverRole", output.Stacks[0].Outputs)
	fmt.Println("EBS CSI Driver role ARN:", roleARN)

	return roleARN, nil
}

func getParameterValueFromOutput(key string, outputs []*cloudformation.Output) string {
	for _, output := range outputs {
		if *output.OutputKey == key {
			return *output.OutputValue
		}
	}

	return ""
}
