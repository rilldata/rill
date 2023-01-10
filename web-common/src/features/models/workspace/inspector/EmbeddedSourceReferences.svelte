<script lang="ts">
  import Shortcut from "@rilldata/web-common/components/tooltip/Shortcut.svelte";
  import TooltipShortcutContainer from "@rilldata/web-common/components/tooltip/TooltipShortcutContainer.svelte";
  import EmbeddedSourceEntry from "@rilldata/web-common/features/sources/embedded/EmbeddedSourceEntry.svelte";
  import { groupURIs } from "@rilldata/web-common/features/sources/group-uris";
  import { formatCompactInteger } from "@rilldata/web-common/lib/formatters";
  import * as classes from "@rilldata/web-local/lib/util/component-classes";
  import { getContext } from "svelte";
  import WithModelResultTooltip from "./WithModelResultTooltip.svelte";

  const queryHighlight = getContext("rill:app:query-highlight");

  export let entries;
  export let modelHasError = false;
  export let references;

  function focus(reference) {
    return () => {
      if (references.length) {
        queryHighlight.set([reference]);
      }
    };
  }
  function blur() {
    queryHighlight.set(undefined);
  }

  $: groupedSources = groupURIs(entries);
</script>

{#each Object.keys(groupedSources) as domain}
  {@const domainSet = groupedSources[domain]}

  {#each domainSet.uris as source (source.name)}
    <WithModelResultTooltip {modelHasError}>
      <a
        href="/source/{source.name}"
        class=" w-full ui-copy-muted flex justify-between
   gap-x-4 {classes.QUERY_REFERENCE_TRIGGER} p-1 pl-4 pr-4"
        on:focus={focus(source)}
        on:mouseover={focus(source)}
        on:mouseleave={blur}
        on:blur={blur}
        class:text-gray-500={modelHasError}
      >
        <EmbeddedSourceEntry
          connector={domainSet.connector}
          path={source.path}
        />

        <div class="text-gray-500 shrink-0">
          {#if source.totalRows}
            {`${formatCompactInteger(source.totalRows)} rows` || ""}
          {/if}
        </div>
      </a>

      <svelte:fragment slot="tooltip-title">
        <div class="break-all">
          {source?.embedded ? source?.source?.properties?.path : source.name}
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
{/each}
