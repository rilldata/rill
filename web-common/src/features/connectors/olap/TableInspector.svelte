<script lang="ts">
  import ColumnProfile from "@rilldata/web-common/features/column-profile/ColumnProfile.svelte";
  import ReconcilingSpinner from "@rilldata/web-common/features/entity-management/ReconcilingSpinner.svelte";
  import CollapsibleSectionTitle from "@rilldata/web-common/layout/CollapsibleSectionTitle.svelte";
  import { LIST_SLIDE_DURATION } from "@rilldata/web-common/layout/config";
  import InspectorHeaderGrid from "@rilldata/web-common/layout/inspector/InspectorHeaderGrid.svelte";
  import { formatInteger } from "@rilldata/web-common/lib/formatters";
  import {
    createQueryServiceTableCardinality,
    createQueryServiceTableColumns,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { slide } from "svelte/transition";

  export let connector: string;
  export let database: string;
  export let databaseSchema: string;
  export let table: string;

  let showColumns = true;
  let isReconciling = false;
  let hasUnsavedChanges = false;

  $: ({ instanceId } = $runtime);

  $: cardinalityQuery = createQueryServiceTableCardinality(instanceId, table, {
    connector,
    database,
    databaseSchema,
  });

  $: profileColumnsQuery = createQueryServiceTableColumns(instanceId, table, {
    connector,
    database,
    databaseSchema,
  });

  $: cardinality = Number($cardinalityQuery?.data?.cardinality ?? 0);

  $: profileColumnsCount =
    $profileColumnsQuery?.data?.profileColumns?.length ?? 0;

  $: rowCount = `${formatInteger(cardinality)} ${
    cardinality !== 1 ? "rows" : "row"
  }`;

  $: columnCount = `${formatInteger(profileColumnsCount)} columns`;
</script>

<div class="wrapper" class:grayscale={hasUnsavedChanges}>
  {#if isReconciling}
    <div class="spinner-wrapper">
      <ReconcilingSpinner />
    </div>
  {:else}
    <InspectorHeaderGrid>
      <slot:fragment slot="top-left">{connector}</slot:fragment>
      <slot:fragment slot="top-right">{rowCount}</slot:fragment>
      <slot:fragment slot="bottom-left">{table}</slot:fragment>
      <slot:fragment slot="bottom-right">{columnCount}</slot:fragment>
    </InspectorHeaderGrid>

    <hr />

    <div>
      <div class="px-4">
        <CollapsibleSectionTitle
          tooltipText="available columns"
          bind:active={showColumns}
        >
          Table columns
        </CollapsibleSectionTitle>
      </div>

      <div transition:slide={{ duration: LIST_SLIDE_DURATION }}>
        <ColumnProfile
          {connector}
          {database}
          {databaseSchema}
          objectName={table}
          indentLevel={0}
        />
      </div>
    </div>
  {/if}
</div>

<style lang="postcss">
  .wrapper {
    @apply transition duration-200 py-2 flex flex-col gap-y-2;
  }

  .spinner-wrapper {
    @apply size-full;
    @apply flex items-center justify-center;
  }
</style>
