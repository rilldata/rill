---
title: Publish your Project to Rill Cloud
sidebar_label: Publish
sidebar_position: 00
---

import TileIcon from '@site/src/components/TileIcon';

<div className="tile-icon-grid">
    <TileIcon
    header="Publish your Dashboard"
    content="Transform and prepare your data with Rill's powerful ETL capabilities."
    link="/build/models/"
    />
    <TileIcon
    header="Configure Deployment Credentials"
    content="Need incremental refreshes or using ClickHouse Modeling? Click here!"
    link="/build/advanced-models"
    />

</div>

Rill Developer is a great tool for building, testing, and viewing your data locally but once your ready to share your findings you'll need to publish the dashboard to Rill Cloud! To understand the differences, see [Rill Developer vs Rill Cloud](/home/concepts/developerVsCloud).

:::tip  first time Publishing?
Publishing your dashboard to Rill for the first time will prompt you to register or login and will automatically start you 30 day free trial! We'll handle all the small setup things that are needed but you can change these at any time.

:::


## Deployment Types

Rill supports two deployment methods for updating your projects:

### GitHub Integration (Recommended)
- **Access Control** - Manage permissions and user access
- **Version History** - Track changes and rollback capabilities  
- **Merge Management** - Review and approve changes through pull requests
- **CI/CD Integration** - Automate deployments through your existing workflow

### UI Button `Update`
- **Direct Updates** - Make changes directly in the Rill interface
- **Immediate Deployment** - Changes are applied instantly

:::tip Recommended Deployment Configuration
We recommend GitHub integration for production environments as it provides better governance and version control. 
:::

## Credentials Management

### Environment Variables
When deploying to Rill Cloud, credentials defined in your local `.env` file are automatically pushed to the cloud environment.

### Connector-Specific Considerations
Some connectors may not dynamically push credentials to your cloud environment:
- **S3** - CLI-based authentication
- **GCS** - Service account configurations  
- **Azure** - CLI-based authentication

### Troubleshooting
If you encounter permission or credential errors:
```bash
rill env configure
```

This command will help resolve authentication issues and ensure proper credential setup.

## Performance Considerations

### DuckDB Performance
When deploying from local development to Rill Cloud, you may notice performance differences with DuckDB-based projects. This is due to:

- **Resource Allocation** - Cloud trial environments have default resource limits
- **Compute Resources** - Different hardware specifications between local and cloud
- **Network Latency** - Additional overhead for cloud-based data processing

### Performance Optimization
If you experience performance degradation:
1. **Contact Support** - Our team can help optimize resource allocation
2. **Resource Scaling** - We can adjust compute resources based on your needs
3. **Query Optimization** - Review and optimize data processing workflows

:::info **Support**
For performance issues or resource allocation concerns, please contact our support team for assistance.
:::