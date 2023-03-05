<script lang="ts">
  import ColumnProfile from "@rilldata/web-common/components/column-profile/ColumnProfile.svelte";
  import { getFilePathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import type { QueryHighlightState } from "@rilldata/web-common/features/models/query-highlight-store";
  import CollapsibleSectionTitle from "@rilldata/web-common/layout/CollapsibleSectionTitle.svelte";
  import { LIST_SLIDE_DURATION } from "@rilldata/web-common/layout/config";
  import {
    useQueryServiceTableCardinality,
    useRuntimeServiceGetCatalogEntry,
    useRuntimeServiceGetFile,
    useRuntimeServiceListCatalogEntries,
  } from "@rilldata/web-common/runtime-client";
  import { getContext } from "svelte";
  import { derived, Writable, writable } from "svelte/store";
  import { slide } from "svelte/transition";
  import { runtime } from "../../../../runtime-client/runtime-store";
  import { getTableReferences } from "../../utils/get-table-references";
  import References from "./References.svelte";
  import { combineEntryWithReference } from "./utils";

  export let modelName: string;

  const queryHighlight: Writable<QueryHighlightState> = getContext(
    "rill:app:query-highlight"
  );

  $: getModel = useRuntimeServiceGetCatalogEntry(
    $runtime?.instanceId,
    modelName
  );
  let entry;
  // refresh entry value only if the data has changed
  $: entry = $getModel?.data?.entry || entry;

  $: getModelFile = useRuntimeServiceGetFile(
    $runtime?.instanceId,
    getFilePathFromNameAndType(modelName, EntityType.Model)
  );

  $: references = getTableReferences($getModelFile?.data.blob ?? "");
  $: getAllSources = useRuntimeServiceListCatalogEntries($runtime?.instanceId, {
    type: "OBJECT_TYPE_SOURCE",
  });

  $: getAllModels = useRuntimeServiceListCatalogEntries($runtime?.instanceId, {
    type: "OBJECT_TYPE_MODEL",
  });

  // for each reference, match to an existing model or source,
  $: referencedThings = derived(
    [getAllSources, getAllModels],
    ([$sources, $models]) => {
      return [
        ...($sources?.data?.entries || []),
        ...($models?.data?.entries || []),
      ]
        ?.filter(combineEntryWithReference(modelName, references))
        ?.map((entry) => {
          // get the reference that matches this entry
          return [
            entry,
            references.find(
              (ref) =>
                ref.reference === entry.name ||
                (entry?.embedded &&
                  entry?.children?.includes(modelName.toLowerCase()))
            ),
          ];
        });
    }
  );

  // associate with the cardinality
  $: referencedWithMetadata = derived(
    $referencedThings.map(([$thing, ref]) => {
      return derived(
        [
          writable($thing),
          writable(ref),
          useQueryServiceTableCardinality($runtime?.instanceId, $thing.name),
        ],
        ([$thing, ref, $cardinality]) => ({
          entry: $thing,
          reference: ref,
          totalRows: +$cardinality?.data?.cardinality,
        })
      );
    }),
    ($referencedThings) => $referencedThings
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
    <References {modelName} />

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
