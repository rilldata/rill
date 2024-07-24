<script lang="ts">
  import {
    ResourceKind,
    useResource,
  } from "../entity-management/resource-selectors";
  import MetricsTableRow from "./MetricsTableRow.svelte";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  const headers = ["Measure", "Definition", "Format", "Description"];

  export let metricsViewName: string;
  export let reorderList: (initIndex: number, newIndex: number) => void;

  $: ({ instanceId } = $runtime);
  $: resourceQuery = useResource(
    instanceId,
    metricsViewName,
    ResourceKind.MetricsView,
  );

  $: ({ data } = $resourceQuery);

  $: measures = data?.metricsView?.spec?.measures || [];

  //   $: console.log($resource.data);
  const gutterWidth = 56;

  let tbody: HTMLTableSectionElement;
</script>

<div class="wrapper">
  <table>
    <colgroup>
      <col style:width="{gutterWidth}px" />
      <!-- <col style:width="{firstColumnWidth}px" />
        <col style:width="{columnWidth}px" />
        {#if $isTimeComparisonActive}
          <col style:width="{columnWidth}px" />
          <col style:width="{columnWidth}px" />
        {:else}
          <col style:width="{columnWidth}px" />
        {/if} -->
    </colgroup>
    <thead>
      <tr>
        <th></th>
        {#each headers as header (header)}
          <th>{header}</th>
        {/each}
      </tr>
    </thead>
    <tbody bind:this={tbody}>
      {#each measures as measure, i (measure.name)}
        <MetricsTableRow {measure} {reorderList} {i} />
      {/each}
    </tbody>
  </table>
</div>

<style lang="postcss">
  thead tr {
    height: 40px;
  }

  th {
    @apply text-left;
    @apply border pl-4 text-slate-500;
  }
  table {
    @apply p-0 m-0 border-spacing-0 border-collapse w-fit;
    @apply font-normal cursor-pointer select-none;
    /* @apply table-fixed; */

    @apply z-10;
    @apply w-full absolute;
  }

  .wrapper {
    @apply border w-full rounded-md;
  }

  tbody {
    @apply bg-gray-100;
  }
</style>
