<script lang="ts">
  import Shortcut from "@rilldata/web-common/components/tooltip/Shortcut.svelte";
  import TooltipShortcutContainer from "@rilldata/web-common/components/tooltip/TooltipShortcutContainer.svelte";
  import EmbeddedSourceEntry from "@rilldata/web-common/features/sources/embedded/EmbeddedSourceEntry.svelte";
  import { formatCompactInteger } from "@rilldata/web-common/lib/formatters";
  import type { V1CatalogEntry } from "@rilldata/web-common/runtime-client";
  import * as classes from "@rilldata/web-local/lib/util/component-classes";
  import { getContext } from "svelte";
  import WithModelResultTooltip from "./WithModelResultTooltip.svelte";

  const queryHighlight = getContext("rill:app:query-highlight");

  export let modelHasError = false;
  export let reference;
  export let entry: V1CatalogEntry;
  export let totalRows: number;

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

<WithModelResultTooltip {modelHasError}>
  <a
    href="/source/{entry.name}"
    class=" w-full ui-copy-muted flex justify-between
   gap-x-4 {classes.QUERY_REFERENCE_TRIGGER} p-1 pl-4 pr-4"
    on:focus={focus(reference)}
    on:mouseover={focus(reference)}
    on:mouseleave={blur}
    on:blur={blur}
    class:text-gray-500={modelHasError}
  >
    <EmbeddedSourceEntry connector={entry.source.connector} path={entry.path} />

    <div class="text-gray-500 shrink-0">
      {#if totalRows}
        {`${formatCompactInteger(totalRows)} rows` || ""}
      {/if}
    </div>
  </a>

  <svelte:fragment slot="tooltip-title">
    <div class="break-all">
      {entry?.embedded ? entry?.source?.properties?.path : entry.name}
    </div></svelte:fragment
  >
  <svelte:fragment slot="tooltip-right">
    {#if entry.source}
      {entry.source.connector}
    {/if}
  </svelte:fragment>

  <svelte:fragment slot="tooltip-description">
    <TooltipShortcutContainer>
      <div>Open in workspace</div>
      <Shortcut>Click</Shortcut>
    </TooltipShortcutContainer>
  </svelte:fragment>
</WithModelResultTooltip>
