---
title: "AWS S3 Bucket"
slug: "aws-s3-bucket"
excerpt: "Batch Ingestion: integrating Rill with S3 storage"
hidden: false
createdAt: "2020-09-11T11:36:27.412Z"
updatedAt: "2021-10-26T21:32:16.983Z"
---
# Setup instructions

## Setup S3 bucket

* Create a new bucket, in the US Standard S3 region, named `rill-${ORG_NAME}-share`, where **"${ORG_NAME}"** is your organizational name. Make sure the bucket name is a DNS-compliant address. For example, the name should not include any underscores; use hyphens instead, as shown in the example above.
* Be sure to disable the "Requester Pays" feature. For more information on S3 bucket name guidelines, see [the AWS documentation](https://docs.aws.amazon.com/AmazonS3/latest/userguide/RequesterPaysBuckets.html).

## Method 1: Bucket ACL Policies

We can provide access to the desired buckets using the bucket policies. You can provide access to the root user for Rill Data AWS Account. 
[block:callout]
{
  "type": "info",
  "body": "arn:aws:iam::248432388601:root",
  "title": "Rill Data AWS Account"
}
[/block]
### Using AWS Console

1. Open S3 bucket console: https://s3.console.aws.amazon.com/ 
1. Choose the S3 bucket to be shared.
1. Select Permissions -> Bucket Policy
1. Add the following bucket policy for the access: (Replace rill-${ORG_NAME}-share or bucket-to-be-shared with the bucket name)


[block:code]
{
  "codes": [
    {
      "code": "{\n   \"Version\": \"2012-10-17\",\n   \"Statement\": [\n       {\n           \"Effect\": \"Allow\",\n           \"Principal\": {\n               \"AWS\": \"arn:aws:iam::248432388601:root\"\n           },\n           \"Action\": [\n               \"s3:GetObject\",\n               \"s3:PutObject\",\n               \"s3:PutObjectAcl\"\n           ],\n           \"Resource\": [\n               \"arn:aws:s3:::rill-${ORG_NAME}-share/*\"\n           ]\n       },\n       {\n           \"Effect\": \"Allow\",\n           \"Principal\": {\n               \"AWS\": \"arn:aws:iam::248432388601:root\"\n           },\n           \"Action\": [\n               \"s3:ListBucket\",\n               \"s3:GetBucketLocation\"\n           ],\n           \"Resource\": [\n               \"arn:aws:s3:::rill-${ORG_NAME}-share\"\n           ]\n       }\n   ]\n}",
      "language": "json"
    }
  ]
}
[/block]

[block:image]
{
  "images": [
    {
      "image": [
        "https://files.readme.io/a2fff1a-Screen_Shot_2020-09-11_at_5.22.23_PM.png",
        "Screen Shot 2020-09-11 at 5.22.23 PM.png",
        1456,
        1866,
        "#dfe1e3"
      ]
    }
  ]
}
[/block]
### Using CLI

We can use CLI to put the bucket policy too. Make sure you have access to put bucket policy.
[block:code]
{
  "codes": [
    {
      "code": "export BUCKET=\"rill-${ORG_NAME}-share\" #Create a bucket with Name of Bucket to be shared\n\ncurl -s http://pkg.rilldata.com/aws/cross-account-s3-bucket-policy.json | sed \"s/bucket-to-be-shared/${BUCKET}/\" | tee policy.json\n\naws s3api put-bucket-policy --bucket ${BUCKET} --policy file://policy.json",
      "language": "shell"
    }
  ]
}
[/block]
## Method 2: AWS IAM Roles

We can provide access to the S3 Bucket through an IAM Role which will be assumed by the Rill Data AWS Account to gain the access to the buckets. 

### Using Cloudformation Console

1. Open AWS Cloudformation to create a new Stack. 
https://console.aws.amazon.com/cloudformation/home?region=us-east-1#/stacks/create/template

2. Use Amazon S3 URL: 
`https://s3.amazonaws.com/cf-templates.rilldata.com/rilldata-s3-bucket-access.yaml`
[block:image]
{
  "images": [
    {
      "image": [
        "https://files.readme.io/6266ce4-Screen_Shot_2020-09-14_at_3.44.05_PM.png",
        "Screen Shot 2020-09-14 at 3.44.05 PM.png",
        2774,
        1534,
        "#e9eaeb"
      ]
    }
  ]
}
[/block]
3. Specify Stack Details
** Stack Name: `rilldata-s3-access`
** Bucket Name: Name of the bucket we want to provide access to.
[block:image]
{
  "images": [
    {
      "image": [
        "https://files.readme.io/01cef67-c2681ce-Screen_Shot_2020-06-03_at_12.46.58_AM.png",
        "c2681ce-Screen_Shot_2020-06-03_at_12.46.58_AM.png",
        2676,
        1490,
        "#f5f5f5"
      ]
    }
  ]
}
[/block]
4. Click Next, Again Next, Acknowledge the Capabilities and Create the Stack.
5. You can check the events and it should create the resources for you.
[block:image]
{
  "images": [
    {
      "image": [
        "https://files.readme.io/209865b-f84093e-Screen_Shot_2020-06-03_at_1.21.47_AM.png",
        "f84093e-Screen_Shot_2020-06-03_at_1.21.47_AM.png",
        3102,
        1200,
        "#e5e7e8"
      ]
    }
  ]
}
[/block]
6. Share the Outputs with Rill Data
[block:image]
{
  "images": [
    {
      "image": [
        "https://files.readme.io/3448175-00dea4a-Screen_Shot_2020-06-03_at_1.31.40_AM.png",
        "00dea4a-Screen_Shot_2020-06-03_at_1.31.40_AM.png",
        2696,
        648,
        "#f4f5f5"
      ]
    }
  ]
}
[/block]
#### CloudFormation Template Reference

We would be using the following Cloudformation Template.
[block:code]
{
  "codes": [
    {
      "code": "AWSTemplateFormatVersion: '2010-09-09'\nMetadata:\n  License: Apache-2.0\n\nDescription: 'AWS CloudFormation Template for providing Rill Data Access to S3 Bucket. It creates a\n  Role that can be assumed by the RillData AWS Account. The Role has a IAM policy associated with them.'\n\nParameters:\n  BucketName:\n    Type: String\n    Description: S3 Bucket Name. Don't append s3://.\n  NamePrefix:\n    Type: String\n    Description: Name prefix for the IAM Policy and IAM Role.\n    Default: rilldata\n  ExternalID:\n    Type: String\n    Description: External ID for Secured Cross Account Access.\n    Default: r!lld@ta\n\nResources:\n  S3Role:\n    Type: AWS::IAM::Role\n    Properties:\n      Description: 'RillData Access to the S3 Bucket. Managed by: Cloudformation'\n      AssumeRolePolicyDocument:\n        Version: '2012-10-17'\n        Statement:\n          - Effect: Allow\n            Principal:\n              AWS:\n                - 'arn:aws:iam::248432388601:root'\n            Action:\n              - 'sts:AssumeRole'\n            Condition:\n              StringEquals:\n                'sts:ExternalId': !Ref ExternalID\n      Policies:\n        - PolicyName: !Join\n            - ''\n            - - !Ref NamePrefix\n              - 'S3AccessPolicy'\n          PolicyDocument:\n            Statement:\n            - Effect: Allow\n              Action: ['s3:*']\n              Resource:\n                - !Join\n                  - ''\n                  - - 'arn:aws:s3:::'\n                    - !Ref BucketName\n                - !Join\n                  - ''\n                  - - 'arn:aws:s3:::'\n                    - !Ref BucketName\n                    - '/*'\n      RoleName: !Join\n        - '-'\n        - - !Ref NamePrefix\n          - s3-access\n      Tags:\n        - Key: Accessor\n          Value: RillData\n        - Key: ManagedBy\n          Value: Cloudformation\n\nOutputs:\n  RoleName:\n    Value: !GetAtt [S3Role, Arn]\n    Description: S3 Access Role Arn, to be shared with RillData\n  ExternalID:\n    Value: !Ref ExternalID\n    Description: ExternalID for Secured Access, to be shared with RillData",
      "language": "yaml"
    }
  ]
}
[/block]
# S3 Logging

As part of your data uploads, Rill Data requires logging to be enabled for the data you intend on sending to us for processing. Rill Data may use these logs for troubleshooting, pipeline optimization, and/or billing. In order to allow Rill Data to access these logs, please follow the instructions below to set up the proper configuration.

If you have not enabled logging for the bucket Rill Data is ingesting data from, please follow these steps:

1. Go to http://aws.amazon.com
    * Login to your AWS account on the AWS Management Console and navigate to the S3 tab
    * Click the “Properties” icon on the upper-right menu to bring up the “Properties” options on the bucket Rill Data is ingesting data from
    * Under “Logging”, click “Enabled”
    * Under “Target Bucket”, select the same name of the bucket Rill Data is processing data from
    * Under “Target Prefix”, make sure “s3logs/” is entered
    * Click Save

2. Log into your Amazon S3 console and open AWS Cloudformation to create a new Stack.
https://console.aws.amazon.com/cloudformation/home?region=us-east-1#/stacks/create/template
*Note: the person implementing this configuration must have permission to manage CloudFormation and IAM resources.*

3. Click on "Create Stack" and under "Choose a template", either upload the template provided by Rill Data, or use the following s3URL: https://s3.amazonaws.com//cf-templates.rilldata.com/RillS3LogRole.yaml

4. The template allows you to set the following parameters:
    * **"Stack name"**
      * Name the stack `rill-${ORG_NAME}-s3Logs`
    * **"S3LogBucketName"**
      * Description: "New or existing bucket for S3 logs"

5. You will then click that you understand that the template will create IAM resources, and then click "Create".

6. Once the stack has finished being created, send your Rill Data representative the results on the "Outputs" tab.
    * **RoleARN**: This IAM role identifier allows Rill Data to ingest the S3 Logs
    * **S3BucketName**: This is the name of S3 Bucket that the Rill Data role can have access to, in this case, it should be the bucket name you inputted into Cloudformation

*Note: If interested, you can set lifecycle policies on the log bucket of interest. You can either leave this variable blank (no lifecycle policy) or 35 days at minimum. This can be applied through the S3 console on the respective bucket. Reach out to your Technical Account Manager or Sales Director if you have any questions.*