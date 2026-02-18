<script lang="ts">
  import { DatabaseIcon } from "lucide-svelte";
  import {
    connectorIconMapping,
    connectorLabelMapping,
  } from "@rilldata/web-common/features/connectors/connector-icon-mapping.ts";

  const PrimaryConnectors = ["clickhouse", "motherduck", "s3", "snowflake"];
  const SecondaryConnectors = ["bigquery", "redshift", "azure"];
</script>

<div class="container">
  <div class="header">
    <DatabaseIcon />
    <span>Connect your data</span>
  </div>

  <div class="primary-connectors">
    {#each PrimaryConnectors as connector (connector)}
      {@const icon = connectorIconMapping[connector]}
      {@const label = connectorLabelMapping[connector] ?? connector}
      <div class="primary-connector-entry">
        <svelte:component this={icon} size="24px" />
        <span>{label}</span>
      </div>
    {/each}
  </div>

  <div class="secondary-connectors">
    {#each SecondaryConnectors as connector (connector)}
      {@const icon = connectorIconMapping[connector]}
      <div class="secondary-connector-entry">
        <svelte:component this={icon} size="24px" />
      </div>
    {/each}
    <span>more</span>
  </div>
</div>

<style lang="postcss">
  .container {
    @apply flex flex-col p-6 gap-4 w-fit;
    @apply border border-primary-200 rounded-lg;
    background: radial-gradient(
      58.72% 82.18% at 23.7% 14.73%,
      #d7e4ff 42.79%,
      #eaecff 96.63%
    );
  }

  /* We need to toggle off hover when primary connector is hovered */
  .container:hover:not(:has(.primary-connector-entry:hover)) {
    @apply border-accent-primary-action shadow-lg cursor-pointer;
  }

  .header {
    @apply flex flex-row items-center;
    @apply text-lg text-fg-primary font-semibold;
  }

  .primary-connectors {
    @apply grid grid-cols-2 gap-3;
  }

  .primary-connector-entry {
    @apply flex flex-row gap-2 items-center p-2 w-40;
    @apply text-sm bg-surface-overlay rounded-md border;
    @apply hover:border-accent-primary-action hover:shadow-lg cursor-pointer;
  }

  .secondary-connectors {
    @apply flex flex-row items-center justify-center gap-3;
  }

  .secondary-connector-entry {
  }
</style>
