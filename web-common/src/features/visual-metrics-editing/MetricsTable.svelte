<script lang="ts">
  import {
    ResourceKind,
    useResource,
  } from "../entity-management/resource-selectors";
  import MetricsTableRow from "./MetricsTableRow.svelte";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  const headers = ["Measure", "Definition", "Format", "Description"];

  export let metricsViewName: string;

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
    <tbody>
      {#each measures as measure (measure.name)}
        <MetricsTableRow {measure} />
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

    @apply w-full;
  }

  .wrapper {
    @apply border w-full rounded-md;
  }
</style>
