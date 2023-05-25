package main

const (
	defaultAudience = "sts.amazonaws.com"

	// CloudFormation templates
	EBSCSIDriverTemplate = `---
AWSTemplateFormatVersion: '2010-09-09'
Description: 'Amazon EKS EBS CSI Driver Role'


Parameters:

  AmazonEBSCSIDriverPolicyArn:
    Type: String
    Default: arn:aws:iam::aws:policy/service-role/AmazonEBSCSIDriverPolicy
    Description: The ARN of the managed policy

Resources:

  AWSEBSCSIDriverRoleForAmazonEKS:
    Type: AWS::IAM::Role
    Properties:
      AssumeRolePolicyDocument:
        Version: '2012-10-17'
        Statement:
        - Effect: Allow
          Principal:
            Federated:
            - !Sub "arn:aws:iam::${AWS::AccountId}:oidc-provider/oidc.eks.{.Region}.amazonaws.com/id/{.ProviderID}"
          Action: sts:AssumeRoleWithWebIdentity
          Condition:
            StringEquals: {
              "oidc.eks.{{.Region}}.amazonaws.com/id/{{.ProviderID}}:sub": "system:serviceaccount:kube-system:ebs-csi-controller-sa",
              "oidc.eks.{{.Region}}.amazonaws.com/id/{{.ProviderID}}:aud": "sts.amazonaws.com"
            }
      Path: "/"
      ManagedPolicyArns:
      - !Ref AmazonEBSCSIDriverPolicyArn

Outputs:

  EBSCSIDriverRole:
    Description: The role that EKS will for enabling the EBS CSI driver
    Value: !GetAtt AWSEBSCSIDriverRoleForAmazonEKS.Arn
    Export:
      Name: !Sub "${AWS::StackName}-RoleArn"

`
)
