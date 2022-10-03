---
title: "Security"
slug: "security"
---
import Excerpt from '@site/src/components/Excerpt'

<Excerpt />

## Credentials Management

We are securing the secrets required for internal communications as well as the database credentials are stored in Hashicorp's Vault. 

## Vault by HashiCorp
Vault is a tool for managing secrets like passwords, access keys, and certificates. Vault allows us to decouple secrets from applications.

## Authentication

### Applications
Vault is authenticated using Kubernetes JWT Auth method. Only application running within the Kubernetes clusters are allowed to access the credentials.

### Operators
Vault Operators are only allowed under strict ACLs and login via OIDC only. Environment specific ACLs have been applied to secure each environment. 

## Secrets Injection

Beyond the basics of securing data in transit and at rest, audit logs, and access controls, Vault Webhook injects the secrets directly into the kubernetes containers bypassing the kubernetes secrets and in etcd.

No confidential data ever persists on the disk - not even temporarily - or in etcd. All secrets are stored in memory, and only visible to the process that requests them.

Additionally, there is no persistent connection with Vault, and any Vault token used to read environment variables is flushed from memory before the application starts, in order to minimize attack surface.