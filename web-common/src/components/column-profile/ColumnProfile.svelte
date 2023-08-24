<script lang="ts">
  import {
    ColumnProfileData,
    getColumnsProfileData,
  } from "@rilldata/web-common/components/column-profile/columns-profile-data";
  import type { ColumnsProfileDataStore } from "@rilldata/web-common/components/column-profile/columns-profile-data";
  import { COLUMN_PROFILE_CONFIG } from "@rilldata/web-common/layout/config";
  import {
    createQueryServiceTableColumns,
    createQueryServiceTableRows,
  } from "@rilldata/web-common/runtime-client";
  import { NATIVE_SELECT } from "@rilldata/web-local/lib/util/component-classes";
  import { onMount } from "svelte";
  import type { Readable } from "svelte/store";
  import { runtime } from "../../runtime-client/runtime-store";
  import { getColumnType } from "./column-types";
  import { defaultSort, sortByName, sortByNullity } from "./utils";

  export let containerWidth = 0;
  export let objectName: string;
  export let indentLevel = 0;

  let mode = "summaries";

  let container;

  onMount(() => {
    const observer = new ResizeObserver(() => {
      containerWidth = container?.clientWidth ?? 0;
    });
    observer.observe(container);
    return () => observer.unobserve(container);
  });

  // get all column profiles.
  let profileColumns;
  $: profileColumns = createQueryServiceTableColumns(
    $runtime?.instanceId,
    objectName,
    {},
    { query: { keepPreviousData: true } }
  );

  /** get single example */
  let exampleValue;
  $: exampleValue = createQueryServiceTableRows(
    $runtime?.instanceId,
    objectName,
    {
      limit: 1,
    }
  );

  let columnsProfile: ColumnsProfileDataStore;
  $: columnsProfile = getColumnsProfileData($runtime?.instanceId, objectName);

  let batchedQuery: Readable<boolean>;
  $: if ($profileColumns) {
    if ($profileColumns?.data && !$profileColumns.isFetching)
      columnsProfile.load($profileColumns);
  }

  $: profile = Object.values($columnsProfile.profiles);
  let sortedProfile: Array<ColumnProfileData>;
  const sortByOriginalOrder = null;

  let sortMethod = defaultSort;
  $: if (profile?.length && sortMethod !== sortByOriginalOrder) {
    sortedProfile = [...profile].sort(sortMethod);
  } else {
    sortedProfile = profile;
  }
</script>

<!-- Dummy read to force rendering -->
{#if $batchedQuery}<div class="hidden" />{/if}

<!-- pl-16 -->
<div
  bind:this={container}
  class="pl-{indentLevel === 1
    ? '10'
    : '4'} pr-5 pb-2 flex justify-between text-gray-500 pt-1"
  class:flex-col={containerWidth < 325}
>
  <select bind:value={sortMethod} class={NATIVE_SELECT} style:font-size="11px">
    <option value={sortByOriginalOrder}>show original order</option>
    <option value={defaultSort}>sort by type</option>
    <option value={sortByNullity}>sort by null %</option>
    <option value={sortByName}>sort by name</option>
  </select>
  <select
    bind:value={mode}
    class={NATIVE_SELECT}
    class:hidden={containerWidth < 325}
    style:font-size="11px"
    style:transform="translateX(4px)"
  >
    <option value="summaries">show summary&nbsp;</option>
    <option value="example">show example</option>
    <option value="hide">hide reference</option>
  </select>
</div>

<div class="pb-4">
  {#if sortedProfile && exampleValue}
    {#each sortedProfile as column (column.name)}
      {@const hideRight = containerWidth < COLUMN_PROFILE_CONFIG.hideRight}
      {@const hideNullPercentage =
        containerWidth < COLUMN_PROFILE_CONFIG.hideNullPercentage}
      {@const compact =
        containerWidth < COLUMN_PROFILE_CONFIG.compactBreakpoint}
      <svelte:component
        this={getColumnType(column.type)}
        type={column.type}
        {objectName}
        columnName={column.name}
        example={$exampleValue?.data?.data?.[0]?.[column.name]}
        store={columnsProfile}
        {mode}
        {hideRight}
        {hideNullPercentage}
        {compact}
      />
    {/each}
  {/if}
</div>
