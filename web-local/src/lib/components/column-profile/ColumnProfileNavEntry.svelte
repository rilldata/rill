<script lang="ts">
  import { onMount } from "svelte";
  import { COLUMN_PROFILE_CONFIG } from "../../application-config";
  import { NATIVE_SELECT } from "../../util/component-classes";
  import ColumnProfile from "./ColumnProfile.svelte";
  import { defaultSort, sortByName, sortByNullity } from "./sort-utils";

  export let containerWidth = 0;

  export let cardinality: number;
  export let profile: any;
  export let head: any; // FIXME
  export let entityId: string;
  export let showContextButton = true;
  export let indentLevel = 0;

  let sortedProfile;
  const sortByOriginalOrder = null;

  let sortMethod = defaultSort;
  $: if (sortMethod !== sortByOriginalOrder) {
    sortedProfile = [...profile].sort(sortMethod);
  } else {
    sortedProfile = profile;
  }

  let previewView = "summaries";

  let container;

  onMount(() => {
    const observer = new ResizeObserver(() => {
      containerWidth = container?.clientWidth ?? 0;
    });
    observer.observe(container);
    return () => observer.unobserve(container);
  });
</script>

<!-- pl-16 -->
<div
  bind:this={container}
  class="pl-{indentLevel === 1
    ? '10'
    : '4'} pr-5 pb-2 flex justify-between text-gray-500"
  class:flex-col={containerWidth < 325}
>
  <select
    style:transform="translateX(-4px)"
    bind:value={sortMethod}
    class={NATIVE_SELECT}
  >
    <option value={sortByOriginalOrder}>show original order</option>
    <option value={defaultSort}>sort by type</option>
    <option value={sortByNullity}>sort by null %</option>
    <option value={sortByName}>sort by name</option>
  </select>
  <select
    style:transform="translateX(4px)"
    bind:value={previewView}
    class={NATIVE_SELECT}
    class:hidden={containerWidth < 325}
  >
    <option value="summaries">show summary&nbsp;</option>
    <option value="example">show example</option>
    <option value="hide">hide reference</option>
  </select>
</div>

<div>
  {#if sortedProfile && head.length}
    {#each sortedProfile as column (column.name)}
      <ColumnProfile
        {indentLevel}
        {entityId}
        example={head[0][column.name] || ""}
        {containerWidth}
        hideNullPercentage={containerWidth <
          COLUMN_PROFILE_CONFIG.hideNullPercentage}
        hideRight={containerWidth < COLUMN_PROFILE_CONFIG.hideRight}
        compactBreakpoint={COLUMN_PROFILE_CONFIG.compactBreakpoint}
        view={previewView}
        name={column.name}
        type={column.type}
        summary={column.summary}
        totalRows={cardinality}
        nullCount={column.nullCount}
      />
    {/each}
  {/if}
</div>
