<script lang="ts">
  import EmbeddedSourceEntry from "@rilldata/web-common/features/sources/embedded/EmbeddedSourceEntry.svelte";
  import { formatCompactInteger } from "@rilldata/web-common/lib/formatters";
  import type { V1CatalogEntry } from "@rilldata/web-common/runtime-client";
  import { getContext } from "svelte";
  import type { Writable } from "svelte/store";
  import type { QueryHighlightState } from "../../models/query-highlight-store";

  const queryHighlight: Writable<QueryHighlightState> = getContext(
    "rill:app:query-highlight"
  );

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

<span
  class=" w-full ui-copy-muted flex justify-between
   gap-x-4 p-1 pl-4 pr-4 hover:bg-yellow-200"
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
</span>
