<script lang="ts">
  import { DatabaseIcon, ArrowRightIcon } from "lucide-svelte";
  import {
    connectorIconMapping,
    connectorLabelMapping,
  } from "@rilldata/web-common/features/connectors/connector-icon-mapping.ts";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { getSupportedTopConnectors } from "@rilldata/web-common/features/add-data/manager/selectors.ts";

  export let startConnectorSelection: (name: string | null) => void = () => {};
  export let onWelcomeScreen = false;

  const runtimeClient = useRuntimeClient();
  const topConnectors = getSupportedTopConnectors(runtimeClient);

  function selectConnector(e: MouseEvent, connector: string) {
    e.preventDefault();
    e.stopPropagation();
    e.stopImmediatePropagation();
    startConnectorSelection(connector);
  }
</script>

<div
  class="container {onWelcomeScreen ? 'container-welcome' : 'container-home'}"
>
  <svelte:element
    this={onWelcomeScreen ? "a" : "button"}
    {...onWelcomeScreen
      ? { href: "/welcome/add-data" }
      : { onclick: () => startConnectorSelection(null) }}
    class="all-connectors"
    aria-label="Connect your data"
  >
    <div class="header">
      <DatabaseIcon class="h-[18px]" />
      <span>Connect your data</span>
    </div>

    <div class="grow"></div>

    <div class="see-more-container" aria-label="See more connectors">
      <span class="grow"></span>
      <div class="see-more">
        <span>See more connectors</span>
        <ArrowRightIcon class="w-4 h-4" />
      </div>
    </div>
  </svelte:element>

  <div class="primary-connectors" role="group">
    {#each $topConnectors as connector (connector)}
      {@const icon = connectorIconMapping[connector]}
      {@const label = connectorLabelMapping[connector] ?? connector}
      <svelte:element
        this={onWelcomeScreen ? "a" : "button"}
        class="primary-connector-entry"
        {...onWelcomeScreen
          ? { href: `/welcome/add-data?schema=${connector}` }
          : { onclick: (e) => selectConnector(e, connector) }}
        aria-label={`Connect to ${connector}`}
      >
        <svelte:component this={icon} />
        <span>{label}</span>
      </svelte:element>
    {/each}
  </div>
</div>

<style lang="postcss">
  .container {
    @apply relative w-96 min-w-96 h-[246px];
  }

  .all-connectors {
    @apply flex flex-col p-6 gap-4 w-full h-full;
    @apply border rounded-lg;
  }

  .container-welcome .all-connectors {
    @apply border-primary-200;
    background: radial-gradient(
      94.8% 95.1% at 23.7% 14.73%,
      #d7e4ff 42.79%,
      #eaecff 96.63%
    );
  }
  :global(.dark) .container-welcome .all-connectors {
    background: radial-gradient(
      94.8% 95.1% at 23.7% 14.73%,
      #31497d 22.12%,
      #3b335f 100%
    );
  }

  .container-home .all-connectors {
    @apply bg-surface-overlay;
  }

  /* We need to toggle off hover when primary connector is hovered */
  .container-welcome .all-connectors:hover {
    @apply border-accent-primary-action shadow-lg cursor-pointer;
  }
  .container-home .all-connectors:hover {
    @apply bg-surface-hover;
  }

  .header {
    @apply flex flex-row items-center gap-1.5;
    @apply text-lg text-fg-primary font-semibold;
  }

  .primary-connectors {
    @apply grid grid-cols-2 gap-3 w-[335px];
    @apply absolute bottom-20 left-6;
  }

  .primary-connector-entry {
    @apply flex flex-row gap-2 items-center px-3 py-2;
    @apply text-sm text-fg-primary bg-surface-overlay rounded-md border;
  }
  .container-welcome .primary-connector-entry:hover {
    @apply border-accent-primary-action shadow-lg cursor-pointer;
  }
  .container-home .primary-connector-entry:hover {
    @apply bg-surface-hover;
  }

  .see-more-container {
    @apply flex flex-row items-center;
  }
  .see-more {
    @apply flex flex-row items-center py-2 gap-1;
    @apply text-sm font-medium text-fg-secondary hover:text-primary;
  }
</style>
