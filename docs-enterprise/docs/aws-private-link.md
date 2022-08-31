---
title: "Private Link"
slug: "aws-private-link"
excerpt: "Share data privately"
hidden: false
createdAt: "2020-09-20T06:09:07.835Z"
updatedAt: "2021-08-11T22:46:57.964Z"
---
The choice between Transit Gateway, VPC peering, and AWS PrivateLink is dependent on connectivity.

**AWS PrivateLink** — Use AWS PrivateLink when you have a client/server set up where you want to allow one or more consumer VPCs unidirectional access to a specific service or set of instances in the service provider VPC. Only the clients in the consumer VPC can initiate a connection to the service in the service provider VPC. This is also a good option when client and servers in the two VPCs have overlapping IP addresses as AWS PrivateLink leverages ENIs within the client VPC such that there are no IP conflicts with the service provider. You can access AWS PrivateLink endpoints over VPC Peering, VPN, and AWS Direct Connect.

**VPC peering and Transit Gateway** — Use VPC peering and Transit Gateway when you want to enable layer-3 IP connectivity between VPCs.

# AWS Private Link

We can create a AWS Network Load Balancer and configure it to serve various services such as Kafka or any other HTTP service.
[block:image]
{
  "images": [
    {
      "image": [
        "https://files.readme.io/aaf59da-VPC_Sharing.png",
        "VPC Sharing.png",
        781,
        411,
        "#f4f0f1"
      ]
    }
  ]
}
[/block]
## Using Cloudformation Console

1. Open AWS Cloudformation to create a new Stack. 
https://console.aws.amazon.com/cloudformation/home?region=us-east-1#/stacks/create/template

2. Use Amazon S3 URL: 
`https://s3.amazonaws.com/cf-templates.rilldata.com/rilldata-private-link.yaml`
[block:image]
{
  "images": [
    {
      "image": [
        "https://files.readme.io/753e584-Screen_Shot_2020-09-22_at_1.01.07_AM.png",
        "Screen Shot 2020-09-22 at 1.01.07 AM.png",
        2616,
        1424,
        "#f3f4f4"
      ]
    }
  ]
}
[/block]
3. Specify Stack Details
  * **Stack Name**: `rilldata-privatelink`
  * **AccountId**: RillData AWS Account ID. 
  * **NlbArn**: Arn of Network Load Balancer (Internal) through which we can share the internal Endpoints
[block:image]
{
  "images": [
    {
      "image": [
        "https://files.readme.io/4f20089-Screen_Shot_2020-09-22_at_12.49.51_AM.png",
        "Screen Shot 2020-09-22 at 12.49.51 AM.png",
        2392,
        1272,
        "#f3f4f4"
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
        "https://files.readme.io/09f200a-Screen_Shot_2020-09-22_at_12.53.03_AM.png",
        "Screen Shot 2020-09-22 at 12.53.03 AM.png",
        3228,
        1338,
        "#f2f4f4"
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
        "https://files.readme.io/3a1fb14-Screen_Shot_2020-09-22_at_1.05.25_AM.png",
        "Screen Shot 2020-09-22 at 1.05.25 AM.png",
        3744,
        822,
        "#f2f3f3"
      ]
    }
  ]
}
[/block]
### CloudFormation Template Reference

We would be using the following Cloudformation Template.
[block:code]
{
  "codes": [
    {
      "code": "AWSTemplateFormatVersion: 2010-09-09\nMetadata:\n  License: Apache-2.0\n\nDescription: 'AWS CloudFormation Template for creating a Private Link for a given Network Load Balancer'\n\nParameters:\n  NlbArn:\n    Type: String\n    Description: ARN of the Network Load Balancer\n    Default: arn:aws:elasticloadbalancing:us-east-1:248432388601:loadbalancer/net/kafka-broker/de46ce872b289b14\n  AccountId:\n    Type: String\n    Description: ID of the account to share the private link with.\n    Default: 417306524257\nResources:\n  EndpointService:\n    Type: AWS::EC2::VPCEndpointService\n    Properties:\n      AcceptanceRequired: True\n      NetworkLoadBalancerArns:\n        - !Ref NlbArn\n  EndpointServicePermissions:\n    Type: AWS::EC2::VPCEndpointServicePermissions\n    Properties:\n      AllowedPrincipals:\n        - !Join\n          - ''\n          - - 'arn:aws:iam::'\n            - !Ref AccountId\n            - ':root'\n      ServiceId: !Ref EndpointService\n\nOutputs:\n  PrivateLinkServiceId:\n    Value: !Ref EndpointService\n    Description: Service ID of the Private Link\n",
      "language": "yaml"
    }
  ]
}
[/block]
# VPC Peering 

A VPC peering connection is a networking connection between two VPCs that enables you to route traffic between them using private IPv4 addresses or IPv6 addresses. Instances in either VPC can communicate with each other as if they are within the same network. You can create a VPC peering connection between your own VPCs, or with a VPC in another AWS account. The VPCs can be in different regions (also known as an inter-region VPC peering connection).
[block:callout]
{
  "type": "warning",
  "title": "Overlapping Subnets",
  "body": "The two VPCs cannot have overlapping IP addresses"
}
[/block]