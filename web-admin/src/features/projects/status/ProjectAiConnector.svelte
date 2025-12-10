<script lang="ts">
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import { createRuntimeServiceGetInstance } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import RillFilled from "@rilldata/web-common/components/icons/RillFilled.svelte";
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

<section class="ai-connector">
  <h3 class="ai-label">AI</h3>
  {#if isLoading}
    <div class="py-1">
      <Spinner status={EntityStatus.Running} size="16px" />
    </div>
  {:else if error}
    <div class="py-0.5">
      <span class="text-red-600">Error loading AI connector</span>
    </div>
  {:else if isUserConfigured}
    <div class="ai-display">
      <div class="ai-header">
        <ConnectorIcon size="16px" />
        <span class="ai-name">
          {getDriverDisplayName(userConnectorConfig?.type)}
        </span>
      </div>
      {#if configDetails.length > 0}
        <div class="ai-details">
          {#each configDetails as detail}
            <div class="ai-detail-row">
              <span class="ai-detail-label">{detail.label}:</span>
              <span class="ai-detail-value">{detail.value}</span>
            </div>
          {/each}
        </div>
      {/if}
    </div>
  {:else}
    <div class="ai-display">
      <div class="ai-header">
        <RillFilled size="16" />
        <span class="ai-name">
          Rill AI
          <span class="ai-default">(default)</span>
        </span>
      </div>
    </div>
  {/if}
</section>

<style lang="postcss">
  .ai-connector {
    @apply flex flex-col gap-y-1;
  }

  .ai-label {
    @apply text-[10px] leading-none font-semibold uppercase;
    @apply text-gray-500;
  }

  .ai-display {
    @apply flex flex-col gap-y-1;
  }

  .ai-header {
    @apply flex items-center gap-x-1.5;
  }

  .ai-name {
    @apply text-[12px] font-semibold text-gray-800;
  }

  .ai-default {
    @apply text-gray-500 font-normal;
  }

  .ai-details {
    @apply flex flex-col gap-y-0.5 pl-6;
  }

  .ai-detail-row {
    @apply flex items-center gap-x-1;
    @apply text-[11px] text-gray-600;
  }

  .ai-detail-label {
    @apply font-mono;
  }

  .ai-detail-value {
    @apply text-gray-800;
  }
</style>
