package main

import (
	"fmt"
	"log"
	"path"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/eks"
)

const (
	ebsCSIAddonName = "aws-ebs-csi-driver"
)

func newEKSCluster() *eks.Cluster {
	clusterName := clusterConfig.Spec.Name
	output, err := clusterConfig.AWS.EKSSservice.DescribeCluster(
		&eks.DescribeClusterInput{
			Name: &clusterName,
		},
	)
	if err != nil {
		log.Fatalf("cannot describe cluster %s: %v", clusterName, err)
	}
	if output == nil {
		log.Fatalf("cluster %s not found", clusterName)
	}
	log.Println("Cluster OIDC provider:", path.Base(*output.Cluster.Identity.Oidc.Issuer))

	return output.Cluster
}

func checkEBSAddon() (string, error) {
	input := eks.DescribeAddonInput{
		AddonName:   aws.String(ebsCSIAddonName),
		ClusterName: aws.String(clusterConfig.Spec.Name),
	}

	output, err := clusterConfig.AWS.EKSSservice.DescribeAddon(&input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			if aerr.Code() == eks.ErrCodeResourceNotFoundException {
				log.Println("EBS CSI driver addon not found: got resource not found exception")
				return "", nil
			}
		}
	}
	if output == nil {
		log.Println("EBS CSI driver addon not found")
		return "", nil
	}
	log.Println("EBS CSI driver addon found:", *output.Addon.AddonArn)

	return *output.Addon.AddonArn, nil
}

func installEBSAddon(roleARN string) error {
	input := eks.CreateAddonInput{
		AddonName:             aws.String(ebsCSIAddonName),
		ClusterName:           aws.String(clusterConfig.Spec.Name),
		ServiceAccountRoleArn: aws.String(roleARN),
	}

	output, err := clusterConfig.AWS.EKSSservice.CreateAddon(&input)
	if err != nil {
		return fmt.Errorf("cannot install EBS CSI driver addon: %v", err)
	}
	fmt.Println("installed addon EBS CSI driver:", *output.Addon.AddonArn)

	return nil
}
