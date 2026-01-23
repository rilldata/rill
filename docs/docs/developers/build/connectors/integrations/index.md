---
title: "Set Up External Integrations"
description: Connect Rill to external services like OpenAI and Slack
sidebar_label: "Integrations"
sidebar_position: 7
---

import ConnectorIcon from '@site/src/components/ConnectorIcon';

External integrations extend Rill's capabilities by connecting to third-party services. Unlike data connectors that import data into Rill, these integrations enable features like AI-powered analytics and notifications.

## Available Integrations

### OpenAI
### Slack


<div className="connector-icon-grid">
  <ConnectorIcon
    icon={<img src="/img/build/connectors/icons/Logo-AI.svg" alt="AI" className="sheets-icon" />}
    header="AI"
    content="Create and define a OpenAI Connector with your own API key."
    link="/developers/build/connectors/integrations/openai"
    linkLabel="Learn more"
    referenceLink="openai"
  />
  <ConnectorIcon
    icon={<img src="/img/build/connectors/icons/Logo-Slack.svg" alt="Slack" className="slack-icon" />}
    header="Slack"
    content="Connect to Slack to send alerts and messages from Rill."
    link="/developers/build/connectors/integrations/slack"
    linkLabel="Learn more"
    referenceLink="slack"
  />
</div>


