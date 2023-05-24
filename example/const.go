package main

const (
	defaultAudience = "sts.amazonaws.com"

	// CloudFormation templates
	EBSCSIDriverTemplate = `---
AWSTemplateFormatVersion: '2010-09-09'
Description: 'Amazon EKS EBS CSI Driver Role'


Parameters:

  OIDCProvider:
    Type: String
    Description: The ID of the OIDC Provider

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
            - !Sub "arn:aws:iam::${AWS::AccountId}:oidc-provider/oidc.eks.${AWS::Region}.amazonaws.com/id/${OIDCProvider}"
          Action: sts:AssumeRoleWithWebIdentity
          Condition:
            StringEquals: 
              Fn::Base64:
                !Sub |
                  oidc.eks.${AWS::Region}.amazonaws.com/id/${OIDCProvider}:sub: system:serviceaccount:kube-system:ebs-csi-controller-sa
      Path: "/"
      ManagedPolicyArns:
      - !Ref AmazonEBSCSIDriverPolicyArn

Outputs:

  RoleArn:
    Description: The role that EKS will for enabling the EBS CSI driver
    Value: !GetAtt AWSEBSCSIDriverRoleForAmazonEKS.Arn
    Export:
      Name: !Sub "${AWS::StackName}-RoleArn"

`
)
