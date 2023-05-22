1. Create IAM OIDC for the cluster with eksctl:
    - Check if cluster is already associated with an OIDC provider
        ```shell
        oidc_id=$(aws eks describe-cluster --name my-cluster --query "cluster.identity.oidc.issuer" --output text | cut -d '/' -f 5)
        aws iam list-open-id-connect-providers | grep $oidc_id | cut -d "/" -f4
        ```
    - Create IAM OIDC if not:
        ```shell
        eksctl utils associate-iam-oidc-provider --cluster my-cluster --approve
        ```

2. Create AWS EBS CSI plugin IAM role: grant permissions to make calls to AWS APIs
    - When the plugin is deployed it is configured to use a service account `ebs-csi-controller-sa`
    - Get the OIDC of the cluster
        ```shell
        aws eks describe-cluster \
          --name my-cluster \
          --query "cluster.identity.oidc.issuer" \
          --output text
        ```
    - Create IAM role with following policy `aws-ebs-csi-driver-trust-policy.json`
        ```json
        {
          "Version": "2012-10-17",
          "Statement": [
            {
              "Effect": "Allow",
              "Principal": {
                "Federated": "arn:aws:iam::{AWS_ACCOUNT_ID}:oidc-provider/oidc.eks.{AWS_REGION}.amazonaws.com/id/{OIDC}"
              },
              "Action": "sts:AssumeRoleWithWebIdentity",
              "Condition": {
                "StringEquals": {
                  "oidc.eks.{AWS_REGION}.amazonaws.com/id/{OIDC}:aud": "sts.amazonaws.com",
                  "oidc.eks.{AWS_REGION}.amazonaws.com/id/{OIDC}:sub": "system:serviceaccount:kube-system:ebs-csi-controller-sa"
                }
              }
            }
          ]
        }
        ```
        ```shell
        aws iam create-role \
          --role-name {ROLE_NAME} \
          --assume-role-policy-document file://"aws-ebs-csi-driver-trust-policy.json"
        ```
        - attach the AWS-managed policy for ebs csi driver
        ```shell
        aws iam attach-role-policy \
          --policy-arn arn:aws:iam::aws:policy/service-role/AmazonEBSCSIDriverPolicy \
          --role-name {ROLE_NAME}
        ```
        - if using encryption for the EBS volumes, add an extra policy with requried KMS actions
        ```json
        kms:CreateGrant,
        kms:ListGrants,
        kms:RevokeGrant
        kms:Encrypt,
        kms:Decrypt,
        kms:ReEncrypt,
        kms:GenerateDataKey,
        kms:DescribeKey,
        ```
        - assign to role

3. Deploy the driver -> 
    ```shell
    kubectl apply -k "github.com/kubernetes-sigs/aws-ebs-csi-driver/deploy/kubernetes/overlays/stable/?ref=release-1.18"
    ```

4. Annotate the `ebs-csi-controller-sa` Kubernetes service account with the ARN of the IAM role
    ```shell
    kubectl annotate serviceaccount ebs-csi-controller-sa \
        -n kube-system \
        eks.amazonaws.com/role-arn=arn:aws:iam::{AWS_ACCOUNT_ID}:role/AmazonEKS_EBS_CSI_DriverRole
    ```
5. Restart the `ebs-csi-controller-sa` controller for the annotation to take effect
    ```shell
    kubectl rollout restart deployment ebs-csi-controller -n kube-system
    ```
