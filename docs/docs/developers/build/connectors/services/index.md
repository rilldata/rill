---
title: "Set Up Service Integrations"
description: Connect Rill to external services like OpenAI and Slack
sidebar_position: 7
---

import ConnectorIcon from '@site/src/components/ConnectorIcon';

Service Integrations extend Rill's capabilities by connecting to third-party services. Unlike data connectors that import data into Rill, these integrations enable features like AI-powered analytics and notifications.

## Available Integrations

### Claude
### OpenAI
### Slack


<div className="connector-icon-grid">
  <ConnectorIcon
    icon={<img src="/img/build/connectors/icons/Logo-Claude.svg" alt="Claude" className="sheets-icon" />}
    header="AI"
    content="Create and define a Claude Connector with your own API key."
    link="/developers/build/connectors/services/claude"
    linkLabel="Learn more"
    referenceLink="claude"
  />
  <ConnectorIcon
    icon={<img src="/img/build/connectors/icons/Logo-AI.svg" alt="OpenAI" className="sheets-icon" />}
    header="AI"
    content="Create and define an OpenAI Connector with your own API key."
    link="/developers/build/connectors/services/openai"
    linkLabel="Learn more"
    referenceLink="openai"
  />
  <ConnectorIcon
    icon={<img src="/img/build/connectors/icons/Logo-Slack.svg" alt="Slack" className="slack-icon" />}
    header="Slack"
    content="Connect to Slack to send alerts and messages from Rill."
    link="/developers/build/connectors/services/slack"
    linkLabel="Learn more"
    referenceLink="slack"
  />
</div>


