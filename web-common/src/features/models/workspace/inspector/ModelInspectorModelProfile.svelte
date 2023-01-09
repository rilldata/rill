<script lang="ts">
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";

  import Shortcut from "@rilldata/web-common/components/tooltip/Shortcut.svelte";
  import TooltipShortcutContainer from "@rilldata/web-common/components/tooltip/TooltipShortcutContainer.svelte";
  import { formatCompactInteger } from "@rilldata/web-common/lib/formatters";
  import {
    useRuntimeServiceGetCatalogEntry,
    useRuntimeServiceGetTableCardinality,
    useRuntimeServiceListCatalogEntries,
    V1CatalogEntry,
  } from "@rilldata/web-common/runtime-client";
  import { LIST_SLIDE_DURATION } from "@rilldata/web-local/lib/application-config";
  import CollapsibleSectionTitle from "@rilldata/web-local/lib/components/CollapsibleSectionTitle.svelte";
  import ColumnProfile from "@rilldata/web-local/lib/components/column-profile/ColumnProfile.svelte";
  import * as classes from "@rilldata/web-local/lib/util/component-classes";
  import { getContext } from "svelte";
  import { derived, writable } from "svelte/store";
  import { slide } from "svelte/transition";
  import { getTableReferences } from "../../utils/get-table-references";
  import EmbeddedSourceReferences from "./EmbeddedSourceReferences.svelte";
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
        return references.some(
          (ref) =>
            ref.reference === entry.name ||
            entry?.embeds?.includes(modelName.toLowerCase())
        );
      })
      .map((entry) => {
        return [
          entry,
          references.find(
            (ref) =>
              ref.reference === entry.name ||
              entry?.embeds?.includes(modelName.toLowerCase())
          ),
        ];
      })
      .map((arr) => {
        const entry = arr[0] as V1CatalogEntry;
        const reference = arr[1];
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

  $: viableEmbeddedSources = $viableSources?.filter((source) => {
    return source?.embeds?.includes(modelName.toLowerCase());
  });

  $: viableExplicitSources = $viableSources?.filter((source) => {
    return !source?.embedded;
  });

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
          tooltipText="Sources"
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
            <EmbeddedSourceReferences
              {references}
              entries={viableEmbeddedSources}
            />
            {#each viableExplicitSources as source (source.name)}
              <WithModelResultTooltip {modelHasError}>
                <a
                  href="/source/{source.name}"
                  class="ui-copy-muted grid justify-between gap-x-2 {classes.QUERY_REFERENCE_TRIGGER} p-1 pl-4 pr-4"
                  style:grid-template-columns="auto max-content"
                  on:focus={focus(source)}
                  on:mouseover={focus(source)}
                  on:mouseleave={blur}
                  on:blur={blur}
                  class:text-gray-500={modelHasError}
                >
                  <div
                    class="text-ellipsis overflow-hidden whitespace-nowrap flex items-center gap-x-2"
                  >
                    <!-- <div class="text-gray-400">
                      {#if source?.embedded}
                        <SourceEmbedded size="13px" />
                      {:else}
                        <Source size="13px" />
                      {/if}
                    </div> -->
                    <div
                      class=" text-ellipsis overflow-hidden whitespace-nowrap"
                    >
                      {source?.embedded
                        ? source?.source?.properties?.path
                        : source.name}
                    </div>
                  </div>

                  <div class="text-gray-500">
                    {#if source.totalRows}
                      {`${formatCompactInteger(source.totalRows)} rows` || ""}
                    {/if}
                  </div>
                </a>

                <svelte:fragment slot="tooltip-title">
                  <div class="break-all">
                    {source?.embedded
                      ? source?.source?.properties?.path
                      : source.name}
                  </div></svelte:fragment
                >
                <svelte:fragment slot="tooltip-right">
                  {#if source.source}
                    {source.source.connector}
                  {/if}
                  <!-- </div>
                  </div> -->
                </svelte:fragment>

                <svelte:fragment slot="tooltip-description">
                  <TooltipShortcutContainer>
                    <div>Open in workspace</div>
                    <Shortcut>Click</Shortcut>
                  </TooltipShortcutContainer>
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
