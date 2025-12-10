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

  // Get the active OLAP connector name (handle empty string as well as null/undefined)
  $: olapConnector = instanceData?.instance?.olapConnector || "duckdb";

  // Check if this is a user-configured connector (exists in projectConnectors)
  $: isUserConfigured = instanceData?.instance?.projectConnectors?.some(
    (c) => c.name === olapConnector,
  );

  // Find the connector to get the driver type
  $: connectorConfig = instanceData?.instance?.projectConnectors?.find(
    (c) => c.name === olapConnector,
  );

  // For user-configured connectors, use the driver type; otherwise use the connector name
  $: driverType = connectorConfig?.type ?? olapConnector;

  // Get display name
  $: displayName = getDisplayName(driverType);

  // Get the icon component for the connector
  $: IconComponent =
    connectorIconMapping[driverType as keyof typeof connectorIconMapping];

  function getDisplayName(driver: string): string {
    const displayNames: Record<string, string> = {
      duckdb: "DuckDB",
      clickhouse: "ClickHouse",
      clickhousecloud: "ClickHouse Cloud",
      druid: "Apache Druid",
      pinot: "Apache Pinot",
      bigquery: "BigQuery",
      snowflake: "Snowflake",
      postgres: "PostgreSQL",
      mysql: "MySQL",
      athena: "Athena",
      redshift: "Redshift",
      motherduck: "MotherDuck",
    };
    return displayNames[driver] ?? driver;
  }
</script>

<div class="config-row">
  <div class="config-label">OLAP Engine</div>
  <div class="config-value">
    {#if isLoading}
      <Spinner status={EntityStatus.Running} size="14px" />
    {:else if error}
      <span class="text-red-600 text-sm">Error loading OLAP connector</span>
    {:else}
      <div class="olap-content">
        {#if isUserConfigured}
          {#if IconComponent}
            <svelte:component this={IconComponent} size="16px" />
          {:else}
            <ConnectorIcon size="16px" />
          {/if}
        {/if}
        <span class="connector-name">
          {#if !isUserConfigured}Rill-managed
          {/if}{displayName}
        </span>
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

  .olap-content {
    @apply flex items-center gap-x-2;
  }

  .connector-name {
    @apply font-medium text-gray-800;
  }
</style>
