<script lang="ts">
  import { getFilePathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import {
    useRuntimeServiceGetCatalogEntry,
    useRuntimeServiceGetFile,
    useRuntimeServiceGetTableCardinality,
    useRuntimeServiceListCatalogEntries,
  } from "@rilldata/web-common/runtime-client";
  import { LIST_SLIDE_DURATION } from "@rilldata/web-local/lib/application-config";
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import CollapsibleSectionTitle from "@rilldata/web-local/lib/components/CollapsibleSectionTitle.svelte";
  import ColumnProfile from "@rilldata/web-local/lib/components/column-profile/ColumnProfile.svelte";
  import { getContext } from "svelte";
  import { derived, writable } from "svelte/store";
  import { slide } from "svelte/transition";
  import { getTableReferences } from "../../utils/get-table-references";
  import References from "./References.svelte";

  export let modelName: string;

  const queryHighlight = getContext("rill:app:query-highlight");

  $: getModel = useRuntimeServiceGetCatalogEntry(
    $runtimeStore?.instanceId,
    modelName
  );
  let entry;
  // refresh entry value only if the data has changed
  $: entry = $getModel?.data?.entry || entry;

  $: getModelFile = useRuntimeServiceGetFile(
    $runtimeStore?.instanceId,
    getFilePathFromNameAndType(modelName, EntityType.Model)
  );

  $: references = getTableReferences($getModelFile?.data.blob ?? "");
  $: getAllSources = useRuntimeServiceListCatalogEntries(
    $runtimeStore?.instanceId,
    { type: "OBJECT_TYPE_SOURCE" }
  );

  $: getAllModels = useRuntimeServiceListCatalogEntries(
    $runtimeStore?.instanceId,
    { type: "OBJECT_TYPE_MODEL" }
  );

  // for each reference, match to an existing model or source,
  $: referencedThings = derived(
    [getAllSources, getAllModels],
    ([$sources, $models]) => {
      return [
        ...($sources?.data?.entries || []),
        ...($models?.data?.entries || []),
      ]
        ?.filter((entry) => {
          // remove entry w/o a matching reference
          return references.some((ref) => {
            return (
              ref.reference === entry.name ||
              entry?.children?.includes(modelName.toLowerCase())
            );
          });
        })
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
          useRuntimeServiceGetTableCardinality(
            $runtimeStore?.instanceId,
            $thing.name
          ),
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
