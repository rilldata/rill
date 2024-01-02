<script lang="ts">
  import Shortcut from "@rilldata/web-common/components/tooltip/Shortcut.svelte";
  import TooltipShortcutContainer from "@rilldata/web-common/components/tooltip/TooltipShortcutContainer.svelte";
  import { getFilePathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import { useModels } from "@rilldata/web-common/features/models/selectors";
  import { useSources } from "@rilldata/web-common/features/sources/selectors";
  import CollapsibleSectionTitle from "@rilldata/web-common/layout/CollapsibleSectionTitle.svelte";
  import { LIST_SLIDE_DURATION } from "@rilldata/web-common/layout/config";
  import { formatCompactInteger } from "@rilldata/web-common/lib/formatters";
  import {
    createQueryServiceTableCardinality,
    createRuntimeServiceGetFile,
  } from "@rilldata/web-common/runtime-client";
  import * as classes from "@rilldata/web-local/lib/util/component-classes";
  import { getContext } from "svelte";
  import { Writable, derived, writable } from "svelte/store";
  import { slide } from "svelte/transition";
  import { runtime } from "../../../../runtime-client/runtime-store";
  import { getTableReferences } from "../../utils/get-table-references";
  import { getMatchingReferencesAndEntries } from "./utils";
  import WithModelResultTooltip from "./WithModelResultTooltip.svelte";
  import type { QueryHighlightState } from "../../query-highlight-store";

  export let modelName: string;

  let showSourceTables = true;
  let modelHasError = false;

  const queryHighlight: Writable<QueryHighlightState | undefined> = getContext(
    "rill:app:query-highlight"
  );

  $: getModelFile = createRuntimeServiceGetFile(
    $runtime?.instanceId,
    getFilePathFromNameAndType(modelName, EntityType.Model)
  );
  $: references = getTableReferences($getModelFile?.data?.blob ?? "");

  $: getAllSources = useSources($runtime?.instanceId);

  $: getAllModels = useModels($runtime?.instanceId);

  // for each reference, match to an existing model or source,
  $: referencedThings = getMatchingReferencesAndEntries(modelName, references, [
    ...($getAllSources?.data ?? []),
    ...($getAllModels?.data ?? []),
  ]);

  // associate with the cardinality
  $: referencedWithMetadata = derived(
    referencedThings.map(([resource, ref]) => {
      return derived(
        [
          writable(resource),
          writable(ref),
          createQueryServiceTableCardinality(
            $runtime?.instanceId,
            resource?.meta?.name?.name ?? ""
          ),
        ],
        ([resource, ref, cardinality]) => ({
          resource,
          reference: ref,
          totalRows: +(cardinality?.data?.cardinality ?? 0),
        })
      );
    }),
    (referencedThings) => referencedThings
  );

  function focus(reference) {
    return () => {
      if (reference) {
        queryHighlight.set([reference]);
      }
    };
  }
  function blur() {
    queryHighlight.set(undefined);
  }
</script>

{#if $referencedWithMetadata?.length}
  <div class="pt-4 pb-4">
    <div class=" pl-4 pr-4">
      <CollapsibleSectionTitle
        tooltipText="References"
        bind:active={showSourceTables}
      >
        Referenced in this model
      </CollapsibleSectionTitle>
    </div>

    {#if showSourceTables}
      <div transition:slide={{ duration: LIST_SLIDE_DURATION }} class="mt-2">
        {#each $referencedWithMetadata as reference}
          <div>
            <WithModelResultTooltip {modelHasError}>
              <a
                href="/{reference?.resource?.source
                  ? 'source'
                  : 'model'}/{reference?.resource?.meta?.name?.name}"
                class="ui-copy-muted grid justify-between gap-x-2 {classes.QUERY_REFERENCE_TRIGGER} pl-4 pr-4"
                style:grid-template-columns="auto max-content"
                on:focus={focus(reference.reference)}
                on:mouseover={focus(reference.reference)}
                on:mouseleave={blur}
                on:blur={blur}
                class:text-gray-500={modelHasError}
              >
                <div class="truncate flex items-center gap-x-2">
                  <div class="truncate">
                    {reference?.resource?.meta?.name?.name}
                  </div>
                </div>

                <div class="text-gray-500">
                  {#if reference?.totalRows}
                    {`${formatCompactInteger(reference.totalRows)} rows` || ""}
                  {/if}
                </div>
              </a>

              <svelte:fragment slot="tooltip-title">
                <div class="break-all">
                  {reference?.resource?.meta?.name?.name}
                </div></svelte:fragment
              >
              <svelte:fragment slot="tooltip-right">
                {#if reference?.resource?.source}
                  {reference?.resource?.source?.state?.connector}
                {/if}
              </svelte:fragment>

              <svelte:fragment slot="tooltip-description">
                <TooltipShortcutContainer>
                  <div>Open in workspace</div>
                  <Shortcut>Click</Shortcut>
                </TooltipShortcutContainer>
              </svelte:fragment>
            </WithModelResultTooltip>
          </div>
        {/each}
      </div>
    {/if}
  </div>
  <hr />
{/if}
