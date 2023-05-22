package main

import (
	"log"
	"path"

	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/aws/aws-sdk-go/service/eks/eksiface"
)

const (
	ebsCSIAddonName = "aws-ebs-csi-driver"
)

func EKS() (eksiface.EKSAPI, error) {
	session, err := createSession(creds.AccessKey, creds.SecretKey, creds.Region)
	if err != nil {
		return nil, err
	}

	return eks.New(session), nil
}

func getEKSCluster() eks.Cluster {
	eksService, err := EKS()
	if err != nil {
		log.Fatalf("cannot start EKS session: %v", err)
	}
	clusterName := "test-carlos"
	output, err := eksService.DescribeCluster(
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

	return eks.Cluster{}
}

func installEKSClusterAddon() {}
