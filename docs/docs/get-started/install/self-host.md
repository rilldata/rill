---
title: How to Self Host Rill
sidebar_label: Self Host Rill
sidebar_position: 10
---
## Self-Hosting Rill

Rill can be deployed and managed in your own infrastructure, giving you complete control over your data, security, and compliance requirements. Self-hosting Rill allows you to run Rill on-premises or in your own cloud environment while maintaining all the features and capabilities of Rill Cloud.

:::tip Getting Started
Self-hosting Rill is available for enterprise customers. To get started with self-hosting, please [contact our team](/contact) to discuss your requirements.
:::

### Benefits of Self-Hosting

- **Data Sovereignty**: Keep your data within your own infrastructure and maintain full control over data residency
- **Security & Compliance**: Meet strict security and compliance requirements by hosting Rill within your own network
- **Custom Infrastructure**: Deploy Rill on your preferred infrastructure stack and integrate with existing systems
- **Cost Control**: Manage compute and storage costs directly without SaaS subscription fees
- **Custom Integrations**: Integrate Rill with your internal systems, authentication providers, and monitoring tools

### Self-Hosting Architecture

When self-hosting Rill, you'll deploy the following components:

- **Rill Runtime**: The core Rill engine that processes data models, metrics views, and serves dashboards
- **OLAP Engine**: Your choice of embedded DuckDB or ClickHouse, or connect to your existing OLAP infrastructure
- **Web UI**: The Rill dashboard interface for exploring data and managing projects
- **API Server**: REST and gRPC APIs for programmatic access and integrations
- **Scheduler**: Automated data refresh and model execution scheduling

### Deployment Options

#### Kubernetes Deployment

Deploy Rill as a containerized application on Kubernetes:

- **Helm Charts**: Use Rill's Helm charts for easy Kubernetes deployment
- **Resource Management**: Configure resource limits and scaling policies
- **High Availability**: Deploy multiple replicas for fault tolerance
- **Ingress Configuration**: Set up ingress controllers for external access

#### Docker Compose

For smaller deployments or development environments:

- **Single-Node Deployment**: Run all Rill components in a Docker Compose stack
- **Easy Setup**: Quick deployment for testing and development
- **Resource Sharing**: Share resources between components on a single host

#### Cloud Provider Services

Deploy Rill on major cloud platforms:

- **AWS**: Deploy on EC2, ECS, or EKS with integration to S3, RDS, and other AWS services
- **Google Cloud**: Run on GCE, GKE, or Cloud Run with BigQuery and GCS integration
- **Azure**: Deploy on Azure VMs, AKS, or Container Instances with Azure Storage integration

### Infrastructure Requirements

#### Compute Resources

- **Minimum**: 4 CPU cores, 16GB RAM for small deployments
- **Recommended**: 8+ CPU cores, 32GB+ RAM for production workloads
- **Scaling**: Horizontal scaling supported for high-availability deployments

#### Storage

- **Project Storage**: Persistent storage for Rill project files and metadata
- **Data Storage**: Integration with your existing data storage (S3, GCS, Azure Blob, etc.)
- **OLAP Storage**: Storage for embedded OLAP engines (DuckDB/ClickHouse)

#### Networking

- **Internal Network**: Communication between Rill components
- **External Access**: Ingress configuration for dashboard access
- **Data Source Access**: Network connectivity to your data sources

### Configuration

Self-hosted Rill deployments are configured through:

- **Environment Variables**: Set runtime configuration and feature flags
- **Configuration Files**: YAML-based configuration for projects and connectors
- **Secrets Management**: Integration with your secrets management system (Vault, AWS Secrets Manager, etc.)

### Monitoring and Observability

Self-hosted deployments include:

- **Metrics Export**: Prometheus-compatible metrics for monitoring
- **Logging**: Structured logging with configurable log levels
- **Tracing**: Distributed tracing support for debugging and performance analysis
- **Health Checks**: Health endpoints for load balancer and monitoring integration

### Security Considerations

- **Authentication**: Integrate with your identity provider (LDAP, SAML, OAuth)
- **Authorization**: Configure role-based access control (RBAC) policies
- **Network Security**: Deploy within your VPC with proper firewall rules
- **Data Encryption**: Encrypt data at rest and in transit
- **Audit Logging**: Comprehensive audit logs for compliance

### Support and Maintenance

Self-hosted deployments include:

- **Updates**: Rolling updates with zero-downtime deployment options
- **Backup & Recovery**: Backup strategies for project metadata and configurations
- **Scaling**: Horizontal and vertical scaling capabilities
- **Troubleshooting**: Diagnostic tools and logging for issue resolution

:::info Interested in Self-Hosting?

Self-hosting Rill is available for enterprise customers. If you're interested in deploying Rill in your own infrastructure, please [contact our team](/contact) to discuss your requirements and get started.

Our team can help you:
- Plan your deployment architecture
- Configure Rill for your infrastructure
- Set up monitoring and observability
- Provide training and support

:::