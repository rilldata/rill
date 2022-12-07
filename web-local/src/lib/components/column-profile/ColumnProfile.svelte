<script lang="ts">
  import {
    useRuntimeServiceGetTableRows,
    useRuntimeServiceProfileColumns,
  } from "@rilldata/web-common/runtime-client";
  import { onMount } from "svelte";
  import { COLUMN_PROFILE_CONFIG } from "../../application-config";
  import { runtimeStore } from "../../application-state-stores/application-store";
  import { NATIVE_SELECT } from "../../util/component-classes";
  import { defaultSort, sortByName, sortByNullity } from "./sort-utils";

  import { getColumnType } from "./column-types";
  import { getSummaries } from "./queries";

  export let containerWidth = 0;
  // const queryClient = useQueryClient();
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

  // function invalidateForModel(queryHash, modelName) {
  //   const r = new RegExp(
  //     `\/v1\/instances\/[a-zA-Z0-9-]+\/queries/[a-zA-Z0-9-]+\/tables\/${modelName}`
  //   );
  //   return r.test(queryHash);
  // }

  // invalidate any existing queries when this key changes.
  // $: if (key) {
  //   queryClient?.resetQueries({
  //     predicate: (query) => {
  //       console.log(
  //         query.queryHash,
  //         invalidateForModel(query.queryHash, objectName)
  //       );
  //       return false;
  //     },
  //   });
  // }

  // get all column profiles.
  let profileColumns;
  $: profileColumns = useRuntimeServiceProfileColumns(
    $runtimeStore?.instanceId,
    objectName,
    {},
    { query: { keepPreviousData: true } }
  );

  /** get single example */
  let exampleValue;
  $: exampleValue = useRuntimeServiceGetTableRows(
    $runtimeStore?.instanceId,
    objectName,
    { limit: 1 }
  );

  let nestedColumnProfileQuery;
  $: if ($profileColumns?.data?.profileColumns) {
    nestedColumnProfileQuery = getSummaries(
      objectName,
      $runtimeStore?.instanceId,
      $profileColumns?.data?.profileColumns
    );
  }

  $: profile = $nestedColumnProfileQuery;
  let sortedProfile;
  const sortByOriginalOrder = null;

  let sortMethod = defaultSort;
  $: if (profile?.length && sortMethod !== sortByOriginalOrder) {
    sortedProfile = [...profile].sort(sortMethod);
  } else {
    sortedProfile = profile;
  }
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
    bind:value={mode}
    class={NATIVE_SELECT}
    class:hidden={containerWidth < 325}
  >
    <option value="summaries">show summary&nbsp;</option>
    <option value="example">show example</option>
    <option value="hide">hide reference</option>
  </select>
</div>

<div>
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
        {mode}
        {hideRight}
        {hideNullPercentage}
        {compact}
      />
    {/each}
  {/if}
</div>
