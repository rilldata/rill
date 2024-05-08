<script lang="ts">
  import Shortcut from "@rilldata/web-common/components/tooltip/Shortcut.svelte";
  import TooltipShortcutContainer from "@rilldata/web-common/components/tooltip/TooltipShortcutContainer.svelte";
  import CollapsibleSectionTitle from "@rilldata/web-common/layout/CollapsibleSectionTitle.svelte";
  import { LIST_SLIDE_DURATION } from "@rilldata/web-common/layout/config";
  import { formatCompactInteger } from "@rilldata/web-common/lib/formatters";
  import {
    V1Resource,
    createQueryServiceTableCardinality,
  } from "@rilldata/web-common/runtime-client";
  import { getContext } from "svelte";
  import { Writable, derived, writable } from "svelte/store";
  import { slide } from "svelte/transition";
  import { runtime } from "../../../../runtime-client/runtime-store";
  import WithModelResultTooltip from "./WithModelResultTooltip.svelte";
  import type { QueryHighlightState } from "../../query-highlight-store";
  import type { Reference } from "../../utils/get-table-references";

  export let referencedThings: [V1Resource, Reference][];
  export let modelHasError: boolean;

  let showSourceTables = true;

  /** classes for elements that trigger the highlight in a model query */
  export const query_reference_trigger =
    "hover:bg-yellow-200 hover:cursor-pointer";

  const queryHighlight: Writable<QueryHighlightState | undefined> = getContext(
    "rill:app:query-highlight",
  );

  $: referencedWithMetadata = derived(
    referencedThings.map(([resource, ref]) => {
      return derived(
        [
          writable(resource),
          writable(ref),
          createQueryServiceTableCardinality(
            $runtime?.instanceId,
            resource?.meta?.name?.name ?? "",
          ),
        ],
        ([resource, ref, cardinality]) => ({
          resource,
          reference: ref,
          totalRows: +(cardinality?.data?.cardinality ?? 0),
        }),
      );
    }),
    (referencedThings) => referencedThings,
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
  <div>
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
                class="ui-copy-muted grid justify-between gap-x-2 {query_reference_trigger} pl-4 pr-4"
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
                </div>
              </svelte:fragment>
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
{/if}
