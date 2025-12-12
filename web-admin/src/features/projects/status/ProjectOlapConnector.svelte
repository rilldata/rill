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

  // Get the active OLAP connector name
  $: olapConnector = instanceData?.instance?.olapConnector ?? "duckdb";

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

  // Get display name (capitalize first letter)
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

<section class="olap-connector">
  <h3 class="olap-label">OLAP Engine</h3>
  {#if isLoading}
    <div class="py-1">
      <Spinner status={EntityStatus.Running} size="16px" />
    </div>
  {:else if error}
    <div class="py-0.5">
      <span class="text-red-600">Error loading OLAP connector</span>
    </div>
  {:else}
    <div class="olap-display">
      {#if IconComponent}
        <svelte:component this={IconComponent} size="16px" />
      {:else}
        <ConnectorIcon size="16px" />
      {/if}
      <span class="olap-name">
        {displayName}
        {#if !isUserConfigured}
          <span class="olap-default">(default)</span>
        {/if}
      </span>
    </div>
  {/if}
</section>

<style lang="postcss">
  .olap-connector {
    @apply flex flex-col gap-y-1;
  }

  .olap-label {
    @apply text-[10px] leading-none font-semibold uppercase;
    @apply text-gray-500;
  }

  .olap-display {
    @apply flex items-center gap-x-1.5;
  }

  .olap-name {
    @apply text-[12px] font-semibold text-gray-800;
  }

  .olap-default {
    @apply text-gray-500 font-normal;
  }
</style>
