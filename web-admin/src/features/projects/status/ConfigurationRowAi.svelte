<script lang="ts">
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import { connectorIconMapping } from "@rilldata/web-common/features/connectors/connector-icon-mapping";
  import { createRuntimeServiceGetInstance } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import ConnectorIcon from "@rilldata/web-common/components/icons/ConnectorIcon.svelte";

  $: ({ instanceId } = $runtime);

  $: instanceQuery = createRuntimeServiceGetInstance(instanceId);
  $: ({ data: instanceData, isLoading, error } = $instanceQuery);

  // Get the active AI connector name
  $: aiConnector = instanceData?.instance?.aiConnector;

  // Check if this is a user-configured connector (exists in projectConnectors)
  $: userConnectorConfig = instanceData?.instance?.projectConnectors?.find(
    (c) => c.name === aiConnector,
  );

  $: isUserConfigured = !!userConnectorConfig;

  // Get the driver type for icon lookup
  $: driverType = userConnectorConfig?.type;

  // Get the icon component for the connector
  $: IconComponent =
    connectorIconMapping[driverType as keyof typeof connectorIconMapping];

  // Extract config properties for user-provided connectors (excluding secrets)
  $: configDetails = getConfigDetails(userConnectorConfig?.config);

  function getConfigDetails(
    config: Record<string, unknown> | undefined,
  ): { label: string; value: string }[] {
    if (!config) return [];

    const details: { label: string; value: string }[] = [];
    const displayOrder = ["model", "base_url", "api_type", "api_version"];

    for (const key of displayOrder) {
      const value = config[key];
      if (value && typeof value === "string" && !isSecretValue(value)) {
        details.push({
          label: getDisplayLabel(key),
          value: value,
        });
      }
    }

    return details;
  }

  function getDisplayLabel(key: string): string {
    const labels: Record<string, string> = {
      model: "Model",
      base_url: "Base URL",
      api_type: "API Type",
      api_version: "API Version",
    };
    return labels[key] ?? key;
  }

  function isSecretValue(value: string): boolean {
    // Don't display templated values or anything that looks like a secret
    return (
      value.includes("{{") ||
      value.includes("}}") ||
      value.toLowerCase().includes("secret") ||
      value.startsWith("sk-")
    );
  }

  function getDriverDisplayName(driver: string | undefined): string {
    if (!driver) return "AI";
    const names: Record<string, string> = {
      openai: "OpenAI",
      admin: "Rill AI",
    };
    return names[driver] ?? driver;
  }
</script>

<div class="config-row">
  <div class="config-label">AI</div>
  <div class="config-value">
    {#if isLoading}
      <Spinner status={EntityStatus.Running} size="14px" />
    {:else if error}
      <span class="text-red-600 text-sm">Error loading AI connector</span>
    {:else if isUserConfigured}
      <div class="ai-content">
        {#if IconComponent}
          <svelte:component this={IconComponent} size="16px" />
        {:else}
          <ConnectorIcon size="16px" />
        {/if}
        <span class="connector-name">
          {getDriverDisplayName(userConnectorConfig?.type)}
        </span>
        {#if configDetails.length > 0}
          {#each configDetails as detail, i}
            <span class="separator">•</span>
            <span class="detail">
              <span class="detail-label">{detail.label}:</span>
              {detail.value}
            </span>
          {/each}
        {/if}
      </div>
    {:else}
      <div class="ai-content">
        <span class="connector-name">Rill-managed</span>
      </div>
    {/if}
  </div>
</div>

<style lang="postcss">
  .config-row {
    @apply flex items-center border-b border-slate-200;
    @apply min-h-[44px];
  }

  .config-row:last-child {
    @apply border-b-0;
  }

  .config-label {
    @apply w-[140px] flex-shrink-0 px-4 py-3;
    @apply text-sm font-medium text-gray-600;
    @apply bg-slate-50;
    @apply border-r border-slate-200;
    @apply whitespace-nowrap;
  }

  .config-value {
    @apply flex-1 px-4 py-3;
    @apply text-sm;
  }

  .ai-content {
    @apply flex items-center gap-x-2 flex-wrap;
  }

  .connector-name {
    @apply font-medium text-gray-800;
  }

  .separator {
    @apply text-gray-400;
  }

  .detail {
    @apply text-gray-600;
  }

  .detail-label {
    @apply text-gray-500;
  }
</style>
