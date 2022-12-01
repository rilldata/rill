<script lang="ts">
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";

  import { useRuntimeServiceGetCatalogEntry } from "@rilldata/web-common/runtime-client";
  import { LIST_SLIDE_DURATION } from "@rilldata/web-local/lib/application-config";
  import CollapsibleSectionTitle from "@rilldata/web-local/lib/components/CollapsibleSectionTitle.svelte";
  import ColumnProfile from "@rilldata/web-local/lib/components/column-profile/ColumnProfile.svelte";
  import { getContext } from "svelte";
  import { slide } from "svelte/transition";
  export let modelName: string;

  const queryHighlight = getContext("rill:app:query-highlight");

  $: getModel = useRuntimeServiceGetCatalogEntry(
    $runtimeStore?.instanceId,
    modelName
  );

  $: entry = $getModel?.data?.entry;

  // get source tables?
  let sourceTableReferences = [];
  let showColumns = true;

  // toggle state for inspector sections
  let showSourceTables = true;

  function focus(reference) {
    return () => {
      // FIXME
      // if (!currentDerivedModel?.error && reference) {
      //   queryHighlight.set(reference.tables);
      // }
    };
  }
  function blur() {
    queryHighlight.set(undefined);
  }

  // FIXME
  let modelHasError = false;
</script>

<div class="model-profile">
  {#if entry && entry?.model?.sql?.trim()?.length}
    <div class="pt-4 pb-4">
      <div class=" pl-4 pr-4">
        <CollapsibleSectionTitle
          tooltipText="sources"
          bind:active={showSourceTables}
        >
          Sources
        </CollapsibleSectionTitle>
      </div>

      {#if showSourceTables}
        <div
          transition:slide|local={{ duration: LIST_SLIDE_DURATION }}
          class="mt-1"
        >
          <!-- FIXME -->
          <!-- {#each sourceTableReferences as table}
            {@const persistentTableRef = $persistentTableStore.entities.find(
              (t) => table.name === t.tableName
            )}
            {@const derivedTableRef = $derivedTableStore.entities.find(
              (derivedTable) => derivedTable?.id === persistentTableRef?.id
            )}
            {@const correspondingTableCardinality =
              derivedTableRef?.cardinality}

            {@const sourceName =
              persistentTableRef?.tableName || "unknown source"}

            {@const sourceIsDefined = !!persistentTableRef?.tableName}

            <WithModelResultTooltip {modelHasError}>
              <div
                class="grid justify-between gap-x-2 {classes.QUERY_REFERENCE_TRIGGER} p-1 pl-4 pr-4"
                style:grid-template-columns="auto max-content"
                on:focus={focus(table)}
                on:mouseover={focus(table)}
                on:mouseleave={blur}
                on:blur={blur}
                class:text-gray-500={modelHasError}
                class:italic={modelHasError}
              >
                <div class="text-ellipsis overflow-hidden whitespace-nowrap">
                  {sourceName}
                </div>

                <div class="text-gray-500 italic">
                  {#if correspondingTableCardinality}
                    {`${formatInteger(correspondingTableCardinality)} rows` ||
                      ""}
                  {/if}
                </div>
              </div>

              <svelte:fragment slot="tooltip-title"
                >{sourceName}</svelte:fragment
              >
              <svelte:fragment slot="tooltip-right">Source</svelte:fragment>

              <svelte:fragment slot="tooltip-description">
                {#if sourceIsDefined}
                  This source table is referenced in the model query.
                {:else}
                  Data source is not known. This is likely due to a source name
                  changing.
                {/if}
              </svelte:fragment>
            </WithModelResultTooltip>
          {/each} -->
        </div>
      {/if}
    </div>

    <hr />

    <div class="pb-4 pt-4">
      <div class=" pl-4 pr-4">
        <CollapsibleSectionTitle
          tooltipText="selected columns"
          bind:active={showColumns}
        >
          selected columns
        </CollapsibleSectionTitle>
      </div>

      {#if showColumns}
        <div transition:slide|local={{ duration: LIST_SLIDE_DURATION }}>
          {#key entry?.model?.sql}
            <ColumnProfile
              key={entry?.model?.sql}
              objectName={entry?.model?.name}
              indentLevel={0}
            />
          {/key}
        </div>
      {/if}
    </div>
  {/if}
</div>
