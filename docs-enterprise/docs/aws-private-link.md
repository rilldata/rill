---
title: "Private Link"
slug: "aws-private-link"
---
import Excerpt from '@site/src/components/Excerpt'

<Excerpt text="Share data privately" />

The choice between Transit Gateway, VPC peering, and AWS PrivateLink is dependent on connectivity.

**AWS PrivateLink** — Use AWS PrivateLink when you have a client/server set up where you want to allow one or more consumer VPCs unidirectional access to a specific service or set of instances in the service provider VPC. Only the clients in the consumer VPC can initiate a connection to the service in the service provider VPC. This is also a good option when client and servers in the two VPCs have overlapping IP addresses as AWS PrivateLink leverages ENIs within the client VPC such that there are no IP conflicts with the service provider. You can access AWS PrivateLink endpoints over VPC Peering, VPN, and AWS Direct Connect.

**VPC peering and Transit Gateway** — Use VPC peering and Transit Gateway when you want to enable layer-3 IP connectivity between VPCs.

## AWS Private Link

We can create a AWS Network Load Balancer and configure it to serve various services such as Kafka or any other HTTP service.
![](https://images.contentful.com/ve6smfzbifwz/45kXT87VvGxRUW5sDRMR33/9203c3ef094fdc0308cca904a5603088/aaf59da-VPC_Sharing.png)
### Using Cloudformation Console

1. Open AWS Cloudformation to create a new Stack. 
https://console.aws.amazon.com/cloudformation/home?region=us-east-1#/stacks/create/template

2. Use Amazon S3 URL: 
`https://s3.amazonaws.com/cf-templates.rilldata.com/rilldata-private-link.yaml`
![](https://images.contentful.com/ve6smfzbifwz/8i5a5zwLQGShNLCwS2n0j/dd5853d08892263cab345434a32fd6d9/753e584-Screen_Shot_2020-09-22_at_1.01.07_AM.png)
3. Specify stack details
  * **Stack Name**: `rilldata-privatelink`
  * **AccountId**: RillData AWS Account ID. 
  * **NlbArn**: Arn of Network Load Balancer (Internal) through which we can share the internal Endpoints
![](https://images.contentful.com/ve6smfzbifwz/4IXumXaYyUTwajFe5OnRHM/12e94687d489971958c3a74210a6fe41/4f20089-Screen_Shot_2020-09-22_at_12.49.51_AM.png)
4. Click Next, Again Next, Acknowledge the Capabilities and Create the Stack.
5. You can check the events and it should create the resources for you.
![](https://images.contentful.com/ve6smfzbifwz/3AxFqm09Q6tPh1dzb72W2k/a4623bf93afb285c61712d94d8b9f5a4/09f200a-Screen_Shot_2020-09-22_at_12.53.03_AM.png)
6. Share the Outputs with Rill Data
![](https://images.contentful.com/ve6smfzbifwz/2TJ6khfcxtMZCz6StFWW8q/14e16933c735731002975bcdabef30a5/3a1fb14-Screen_Shot_2020-09-22_at_1.05.25_AM.png)
#### CloudFormation Template Reference

We would be using the following Cloudformation Template.

```yaml title="YAML"
AWSTemplateFormatVersion: 2010-09-09
Metadata:
  License: Apache-2.0

Description: 'AWS CloudFormation Template for creating a Private Link for a given Network Load Balancer'

Parameters:
  NlbArn:
    Type: String
    Description: ARN of the Network Load Balancer
    Default: arn:aws:elasticloadbalancing:us-east-1:248432388601:loadbalancer/net/kafka-broker/de46ce872b289b14
  AccountId:
    Type: String
    Description: ID of the account to share the private link with.
    Default: 417306524257
Resources:
  EndpointService:
    Type: AWS::EC2::VPCEndpointService
    Properties:
      AcceptanceRequired: True
      NetworkLoadBalancerArns:
        - !Ref NlbArn
  EndpointServicePermissions:
    Type: AWS::EC2::VPCEndpointServicePermissions
    Properties:
      AllowedPrincipals:
        - !Join
          - ''
          - - 'arn:aws:iam::'
            - !Ref AccountId
            - ':root'
      ServiceId: !Ref EndpointService

Outputs:
  PrivateLinkServiceId:
    Value: !Ref EndpointService
    Description: Service ID of the Private Link
```

## VPC Peering 

A VPC peering connection is a networking connection between two VPCs that enables you to route traffic between them using private IPv4 addresses or IPv6 addresses. Instances in either VPC can communicate with each other as if they are within the same network. You can create a VPC peering connection between your own VPCs, or with a VPC in another AWS account. The VPCs can be in different regions (also known as an inter-region VPC peering connection).

:::caution Overlapping Subnets
The two VPCs cannot have overlapping IP addresses
:::
