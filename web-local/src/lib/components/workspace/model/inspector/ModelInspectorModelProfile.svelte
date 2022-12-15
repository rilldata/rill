<script lang="ts">
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";

  import {
    useRuntimeServiceGetCatalogEntry,
    useRuntimeServiceGetTableCardinality,
    useRuntimeServiceListCatalogEntries,
  } from "@rilldata/web-common/runtime-client";
  import { LIST_SLIDE_DURATION } from "@rilldata/web-local/lib/application-config";
  import CollapsibleSectionTitle from "@rilldata/web-local/lib/components/CollapsibleSectionTitle.svelte";
  import ColumnProfile from "@rilldata/web-local/lib/components/column-profile/ColumnProfile.svelte";
  import * as classes from "@rilldata/web-local/lib/util/component-classes";
  import { formatInteger } from "@rilldata/web-local/lib/util/formatters";
  import { getTableReferences } from "@rilldata/web-local/lib/util/get-table-references";
  import { getContext } from "svelte";
  import { derived, writable } from "svelte/store";
  import { slide } from "svelte/transition";
  import WithModelResultTooltip from "./WithModelResultTooltip.svelte";
  export let modelName: string;

  const queryHighlight = getContext("rill:app:query-highlight");

  $: getModel = useRuntimeServiceGetCatalogEntry(
    $runtimeStore?.instanceId,
    modelName
  );
  let entry;
  // refresh entry value only if the data has changed
  $: entry = $getModel?.data?.entry || entry;

  $: references = getTableReferences(entry?.model?.sql);
  $: getAllSources = useRuntimeServiceListCatalogEntries(
    $runtimeStore?.instanceId,
    { type: "OBJECT_TYPE_SOURCE" }
  );

  $: viableSources = derived(
    $getAllSources?.data?.entries
      ?.filter((entry) => {
        return references.some((ref) => ref.reference === entry.name);
      })
      .map((entry) => {
        return [entry, references.find((ref) => ref.reference === entry.name)];
      })
      .map(([entry, reference]) => {
        return derived(
          [
            writable(entry),
            writable(reference),
            useRuntimeServiceGetTableCardinality(
              $runtimeStore?.instanceId,
              entry.name
            ),
          ],
          ([entry, reference, $cardinality]) => {
            return {
              ...entry,
              ...reference,
              totalRows: +$cardinality?.data?.cardinality,
            };
          }
        );
      }),
    ($row) => $row
  );
  let showColumns = true;

  // toggle state for inspector sections
  let showSourceTables = true;

  function focus(reference) {
    return () => {
      // FIXME
      if (references.length) {
        queryHighlight.set([reference]);
      }
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
          {#if viableSources && $viableSources}
            {#each $viableSources as table (table.name)}
              <WithModelResultTooltip {modelHasError}>
                <div
                  class="grid justify-between gap-x-2 {classes.QUERY_REFERENCE_TRIGGER} p-1 pl-4 pr-4"
                  style:grid-template-columns="auto max-content"
                  on:focus={focus(table)}
                  on:mouseover={focus(table)}
                  on:mouseleave={blur}
                  on:blur={blur}
                  class:text-gray-500={modelHasError}
                >
                  <div class="text-ellipsis overflow-hidden whitespace-nowrap">
                    {table.name}
                  </div>

                  <div class="text-gray-500">
                    {#if table.totalRows}
                      {`${formatInteger(table.totalRows)} rows` || ""}
                    {/if}
                  </div>
                </div>

                <svelte:fragment slot="tooltip-title"
                  >{table.name}</svelte:fragment
                >
                <svelte:fragment slot="tooltip-right">Source</svelte:fragment>

                <svelte:fragment slot="tooltip-description">
                  This source table is referenced in the model query
                </svelte:fragment>
              </WithModelResultTooltip>
            {/each}
          {/if}
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
          Selected columns
        </CollapsibleSectionTitle>
      </div>

      {#if showColumns}
        <div transition:slide|local={{ duration: LIST_SLIDE_DURATION }}>
          <!-- {#key entry?.model?.sql} -->
          <ColumnProfile
            key={entry?.model?.sql}
            objectName={entry?.model?.name}
            indentLevel={0}
          />
          <!-- {/key} -->
        </div>
      {/if}
    </div>
  {/if}
</div>
