<script lang="ts">
  import { Button } from "../../../../components/button";
  import {
    CLICKHOUSE_SOURCE_CONNECTORS,
    DUCKDB_SOURCE_CONNECTORS,
  } from "../../../connectors/connector-availability";
  import { logoIconMapping } from "../../../connectors/connector-icon-mapping";
  import type { OlapDriver } from "../../../connectors/olap/olap-config";

  export let olapDriver: OlapDriver;
  export let firstDataSource: string | undefined;
  export let onSelectFirstDataSource: (dataSource: string) => void;
  export let onContinue: () => void;
  export let onSkip: () => void;

  $: dataSources =
    olapDriver === "duckdb"
      ? DUCKDB_SOURCE_CONNECTORS
      : CLICKHOUSE_SOURCE_CONNECTORS;
</script>

<div class="data-sources">
  <h2 class="text-subheading">Choose a first data source to add.</h2>
  <div class="source-grid">
    {#each dataSources as source (source)}
      <button
        class="source-button"
        class:active={firstDataSource === source}
        on:click={() => onSelectFirstDataSource(source)}
      >
        <svelte:component this={logoIconMapping[source]} />
      </button>
    {/each}
  </div>
</div>

{#if firstDataSource}
  <Button wide type="primary" on:click={onContinue}>Continue</Button>
{:else}
  <Button wide type="secondary" on:click={onSkip}>Skip</Button>
{/if}

<!-- <div class="help-text">
      Don't see what you're looking for? <a href="#">Request a new connector</a>
    </div> -->

<style lang="postcss">
  .data-sources {
    @apply pt-6;
    @apply flex flex-col gap-y-2;
  }

  .source-grid {
    @apply grid;
    @apply grid-cols-[repeat(5,160px)];
    @apply gap-2;
    @apply my-8;
    @apply justify-center;
  }

  .source-button {
    @apply p-4;
    @apply border border-slate-200;
    @apply rounded-lg;
    @apply flex flex-col items-center;
    @apply gap-2;
    @apply cursor-pointer;
    @apply w-40 h-20;
  }

  .source-button:hover {
    @apply bg-slate-50;
  }

  .source-button.active {
    @apply border-2 border-primary-300 bg-slate-50;
  }
</style>
