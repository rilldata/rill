<script lang="ts">
  import ColumnProfile from "$lib/components/column-profile/ColumnProfile.svelte";
  import { COLUMN_PROFILE_CONFIG } from "$lib/application-config";
  import { NATIVE_SELECT } from "$lib/util/component-classes";
  import { defaultSort } from "$lib/components/column-profile/sort-utils";
  import {
    sortByName,
    sortByNullity,
  } from "$lib/components/column-profile/sort-utils.js";
  import Spacer from "$lib/components/icons/Spacer.svelte";

  export let containerWidth = 0;

  export let cardinality: number;
  export let profile: any;
  export let head: any; // FIXME
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
</script>

<!-- pl-16 -->
<div
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
  </select>
</div>

<div>
  {#if sortedProfile && sortedProfile.length && head.length}
    {#each sortedProfile as column (column.name)}
      <ColumnProfile
        {indentLevel}
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
      >
        <button slot="context-button" class:hidden={!showContextButton}>
          <Spacer size="16px" />
        </button>
      </ColumnProfile>
    {/each}
  {/if}
</div>
