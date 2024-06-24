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

  $: connectors = createRuntimeServiceAnalyzeConnectors($runtime.instanceId, {
    query: {
      // sort alphabetically
      select: (data) => {
        if (!data?.connectors) return;
        const connectors = data.connectors.sort((a, b) =>
          (a?.name as string).localeCompare(b?.name as string),
        );
        return { connectors };
      },
    },
  });
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
    <div class="wrapper">
      {#if error}
        <span class="message">
          {error.message}
        </span>
      {:else if data?.connectors}
        {#if data.connectors.length === 0}
          <span class="message"
            >No connectors found. Add data to get started!</span
          >
        {:else}
          <ol transition:slide={{ duration }}>
            {#each data.connectors as connector (connector.name)}
              <ConnectorEntry {connector} />
            {/each}
          </ol>
        {/if}
      {/if}
    </div>
  {/if}
</section>

<style lang="postcss">
  section {
    @apply flex flex-col relative;
    @apply border-t border-t-gray-200;
  }

  button {
    @apply flex justify-between items-center w-full;
    @apply pl-2 pr-3.5 py-2;
    @apply text-gray-500;
  }

  button:hover {
    @apply bg-slate-100;
  }

  h3 {
    @apply font-semibold text-[10px] uppercase;
  }

  .wrapper {
    @apply overflow-auto;
  }

  .message {
    @apply pl-2 pr-3.5 py-2;
    @apply flex;
    @apply text-gray-500;
  }
</style>
