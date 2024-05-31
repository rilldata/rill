<script lang="ts">
  import { slide } from "svelte/transition";
  import CaretDownIcon from "../../components/icons/CaretDownIcon.svelte";
  import Resizer from "../../layout/Resizer.svelte";
  import { LIST_SLIDE_DURATION as duration } from "../../layout/config";
  import { createRuntimeServiceAnalyzeConnectors } from "../../runtime-client";
  import { runtime } from "../../runtime-client/runtime-store";
  import ConnectorEntry from "./ConnectorEntry.svelte";

  export let startingHeight: number;

  const MIN_HEIGHT = 43; // The height of the "Connectors" header

  let showConnectors = true;
  let sectionHeight = startingHeight;

  $: connectors = createRuntimeServiceAnalyzeConnectors($runtime.instanceId);
  $: ({ data, error } = $connectors);
</script>

<section style:min-height="{MIN_HEIGHT}px" style:height="{sectionHeight}px">
  <Resizer
    bind:dimension={sectionHeight}
    direction="NS"
    side="top"
    min={10}
    basis={showConnectors ? startingHeight : MIN_HEIGHT}
    max={2000}
  />
  <button
    on:click={() => {
      showConnectors = !showConnectors;
    }}
  >
    <h3>Connectors</h3>
    <CaretDownIcon
      className="transform transition-transform {showConnectors
        ? 'rotate-0'
        : '-rotate-180'}"
    />
  </button>
  {#if showConnectors}
    {#if error}
      <span class="message">
        {error.message}
      </span>
    {:else if data?.connectors}
      {#if data.connectors.length === 0}
        <span class="message">No connectors found</span>
      {:else}
        <ol transition:slide={{ duration }}>
          {#each data.connectors as connector (connector.name)}
            <ConnectorEntry {connector} />
          {/each}
        </ol>
      {/if}
    {/if}
  {/if}
</section>

<style lang="postcss">
  section {
    @apply flex flex-col border-t border-t-gray-200 relative;
  }

  button {
    @apply flex justify-between items-center w-full pl-2 pr-3.5 pt-2 pb-2 text-gray-500;
  }

  button:hover {
    @apply bg-gray-200;
  }

  h3 {
    @apply font-semibold text-[10px] uppercase;
  }

  ol {
    @apply flex flex-col;
  }

  .message {
    @apply pl-2 pr-3.5 pt-2 pb-2 text-gray-500;
  }
</style>
