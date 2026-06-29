<script lang="ts">
  import ChevronRight from "@rilldata/web-common/components/icons/ChevronRight.svelte";
  import ExternalLink from "@rilldata/web-common/components/icons/ExternalLink.svelte";
  import Spacer from "@rilldata/web-common/components/icons/Spacer.svelte";
  import { LOADING_CELL } from "@rilldata/web-common/features/dashboards/pivot/pivot-constants";
  import * as m from "@rilldata/web-common/paraglide/messages.js";
  import type { Row } from "tanstack-table-8-svelte-5";
  import type { PivotDataRow } from "./types";

  export let row: Row<PivotDataRow>;
  export let value: string;
  export let assembled = true;
  export let hasNestedDimensions = false;
  // When set, renders a hover-revealed external link icon for URI dimensions.
  export let href: string | undefined = undefined;
  // Flat tables reuse this component only to render the value (and optional
  // link); they must not show the expand chevron or nesting indentation.
  export let expandable = true;

  $: canExpand = expandable && row.getCanExpand();
  $: expanded = row.getIsExpanded();
  $: assembledAndCanExpand = assembled && canExpand;

  $: needsSpacer =
    expandable && (row.depth >= 1 || (hasNestedDimensions && !canExpand));

  function handleExpandClick(e: MouseEvent) {
    e.stopPropagation();
    if (assembledAndCanExpand) {
      row.getToggleExpandedHandler()();
    }
  }
</script>

<div
  role="presentation"
  class="dimension-cell"
  style:padding-left="{row.depth * 14}px"
>
  {#if value === LOADING_CELL}
    <span class="loading-cell"></span>
  {:else if assembledAndCanExpand}
    <button
      type="button"
      tabindex="-1"
      aria-label={expanded ? m.pivot_collapse_row() : m.pivot_expand_row()}
      class="caret opacity-100 shrink-0 cursor-pointer"
      class:expanded
      onclick={handleExpandClick}
    >
      <ChevronRight size="16px" />
    </button>
  {:else if needsSpacer}
    <span class="shrink-0"><Spacer size="16px" /></span>
  {/if}

  <span class="truncate min-w-0">
    {#if value === LOADING_CELL}
      {""}
    {:else if value === ""}
      {"\u00A0"}
    {:else}
      {value ?? "null"}
    {/if}
  </span>

  {#if href}
    <a
      class="external-link shrink-0"
      target="_blank"
      rel="noopener noreferrer"
      {href}
      title={href}
      onclick={(e) => e.stopPropagation()}
    >
      <ExternalLink className="fill-primary-600" />
    </a>
  {/if}
</div>

<style lang="postcss">
  .loading-cell {
    @apply h-2 bg-gray-200 rounded w-full inline-block;
  }

  .dimension-cell {
    @apply flex items-center gap-x-1;
  }

  .external-link {
    @apply inline-flex items-center justify-center transition-opacity;
    opacity: 0;
  }

  .dimension-cell:hover .external-link {
    opacity: 0.7;
  }

  .caret {
    @apply grid size-4 place-items-center rounded-sm border-0 bg-transparent p-0 text-gray-400 transition-colors;
    @apply hover:bg-surface-active hover:text-fg-primary;
    @apply focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-primary-400;
  }

  .caret.expanded {
    @apply opacity-100 transform rotate-90;
  }
</style>
