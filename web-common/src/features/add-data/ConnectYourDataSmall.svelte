<script lang="ts">
  import { DatabaseIcon } from "lucide-svelte";
  import {
    connectorIconMapping,
    connectorLabelMapping,
  } from "@rilldata/web-common/features/connectors/connector-icon-mapping.ts";

  export let startConnectorSelection: (name: string | null) => void;
  export let onWelcomeScreen = false;

  const PrimaryConnectors = ["clickhouse", "motherduck", "s3", "snowflake"];
  const SecondaryConnectors = ["bigquery", "redshift", "azure"];

  let suppressJitter = false;
  let suppressJitterTimeout: ReturnType<typeof setTimeout> | null = null;

  function handleSuppressJitter() {
    suppressJitter = true;
    if (suppressJitterTimeout) clearTimeout(suppressJitterTimeout);
    suppressJitterTimeout = setTimeout(clearSuppressJitter, 250);
  }
  function clearSuppressJitter() {
    suppressJitter = false;
    if (suppressJitterTimeout) clearTimeout(suppressJitterTimeout);
    suppressJitterTimeout = null;
  }

  function selectConnector(e: MouseEvent, connector: string) {
    e.preventDefault();
    e.stopPropagation();
    e.stopImmediatePropagation();
    startConnectorSelection(connector);
  }
</script>

<button
  class="container {onWelcomeScreen ? 'container-welcome' : 'container-home'}"
  on:click={() => startConnectorSelection(null)}
  class:jitter-suppress={suppressJitter}
  aria-label="Connect your data"
>
  <div class="header">
    <DatabaseIcon />
    <span>Connect your data</span>
  </div>

  <div
    class="primary-connectors"
    on:mouseleave={clearSuppressJitter}
    role="group"
  >
    {#each PrimaryConnectors as connector (connector)}
      {@const icon = connectorIconMapping[connector]}
      {@const label = connectorLabelMapping[connector] ?? connector}
      <button
        class="primary-connector-entry"
        on:click={(e) => selectConnector(e, connector)}
        on:mouseleave={handleSuppressJitter}
        aria-label={`Connect to ${connector}`}
      >
        <svelte:component this={icon} />
        <span>{label}</span>
      </button>
    {/each}
  </div>

  <div class="secondary-connectors">
    {#each SecondaryConnectors as connector (connector)}
      {@const icon = connectorIconMapping[connector]}
      <!-- Note that these are not clickable as per design. It is meant to be a preview only -->
      <div class="secondary-connector-entry">
        <svelte:component this={icon} size="24px" />
      </div>
    {/each}
    <span>more</span>
  </div>
</button>

<style lang="postcss">
  .container {
    @apply flex flex-col p-6 gap-4 w-fit;
    @apply border rounded-lg;
  }

  .container-welcome {
    @apply border-primary-200;
    background: radial-gradient(
      58.72% 82.18% at 23.7% 14.73%,
      #d7e4ff 42.79%,
      #eaecff 96.63%
    );
  }

  .container-home {
    @apply bg-surface-overlay;
  }

  /* We need to toggle off hover when primary connector is hovered */
  .container-welcome:hover:not(:has(.primary-connector-entry:hover)):not(
      .jitter-suppress
    ) {
    @apply border-accent-primary-action shadow-lg cursor-pointer;
  }
  .container-home:hover:not(:has(.primary-connector-entry:hover)):not(
      .jitter-suppress
    ) {
    @apply bg-surface-hover;
  }

  .header {
    @apply flex flex-row items-center gap-2;
    @apply text-lg text-fg-primary font-semibold;
  }

  .primary-connectors {
    @apply grid grid-cols-2 gap-3;
  }

  .primary-connector-entry {
    @apply flex flex-row gap-2 items-center p-2 w-40;
    @apply text-sm bg-surface-overlay rounded-md border;
  }
  .container-welcome .primary-connector-entry:hover {
    @apply border-accent-primary-action shadow-lg cursor-pointer;
  }
  .container-home .primary-connector-entry:hover {
    @apply bg-surface-hover;
  }

  .secondary-connectors {
    @apply flex flex-row items-center justify-center gap-3;
  }

  .secondary-connector-entry {
  }
</style>
