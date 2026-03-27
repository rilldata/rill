<script lang="ts">
  import { DatabaseIcon } from "lucide-svelte";
  import {
    connectorIconMapping,
    connectorLabelMapping,
  } from "@rilldata/web-common/features/connectors/connector-icon-mapping.ts";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { getSupportedTopConnectors } from "@rilldata/web-common/features/add-data/manager/selectors.ts";

  export let startConnectorSelection: (name: string | null) => void;
  export let onWelcomeScreen = false;

  const runtimeClient = useRuntimeClient();
  const topConnectors = getSupportedTopConnectors(runtimeClient);

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
  onclick={() => startConnectorSelection(null)}
  class:jitter-suppress={suppressJitter}
  aria-label="Connect your data"
>
  <div class="header">
    <DatabaseIcon />
    <span>Connect your data</span>
  </div>

  <div
    class="primary-connectors"
    onmouseleave={clearSuppressJitter}
    role="group"
  >
    {#each $topConnectors as connector (connector)}
      {@const icon = connectorIconMapping[connector]}
      {@const label = connectorLabelMapping[connector] ?? connector}
      <div
        class="primary-connector-entry"
        onclick={(e) => selectConnector(e, connector)}
        onkeydown={(e) => {
          if (e.key === "Enter" || e.key === " ") {
            e.preventDefault();
            startConnectorSelection(connector);
          }
        }}
        onmouseleave={handleSuppressJitter}
        aria-label={`Connect to ${connector}`}
        role="button"
        tabindex="-1"
      >
        <svelte:component this={icon} />
        <span>{label}</span>
      </div>
    {/each}
  </div>

  <div class="see-all">See all</div>
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
  :global(.dark) .container-welcome {
    background: radial-gradient(
      94.8% 95.1% at 23.7% 14.73%,
      #31497d 22.12%,
      #3b335f 100%
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

  .see-all {
    @apply text-xs text-fg-secondary hover:text-primary;
  }
</style>
