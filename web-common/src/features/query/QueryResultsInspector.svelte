<script lang="ts">
  import DataTypeIcon from "@rilldata/web-common/components/data-types/DataTypeIcon.svelte";
  import CollapsibleSectionTitle from "@rilldata/web-common/layout/CollapsibleSectionTitle.svelte";
  import Inspector from "@rilldata/web-common/layout/workspace/Inspector.svelte";
  import { formatInteger } from "@rilldata/web-common/lib/formatters";
  import type { V1StructType } from "@rilldata/web-common/runtime-client";
  import { slide } from "svelte/transition";
  import { LIST_SLIDE_DURATION } from "../../layout/config";

  export let filePath: string;
  export let schema: V1StructType | null;
  export let rowCount: number;
  export let executionTimeMs: number | null;

  let showColumns = true;

  $: fields = schema?.fields ?? [];
  $: columnCount = fields.length;
</script>

<Inspector {filePath}>
  <div class="py-2 flex flex-col gap-y-2">
    {#if schema}
      <div class="px-4 grid grid-cols-2 gap-y-1 text-xs">
        <span class="text-fg-secondary">Rows</span>
        <span class="text-right font-medium">{formatInteger(rowCount)}</span>

        <span class="text-fg-secondary">Columns</span>
        <span class="text-right font-medium">{formatInteger(columnCount)}</span>

        {#if executionTimeMs !== null}
          <span class="text-fg-secondary">Time</span>
          <span class="text-right font-medium">
            {executionTimeMs < 1000
              ? `${executionTimeMs}ms`
              : `${(executionTimeMs / 1000).toFixed(1)}s`}
          </span>
        {/if}
      </div>

      <hr />

      <div>
        <div class="px-4">
          <CollapsibleSectionTitle
            tooltipText="result columns"
            bind:active={showColumns}
          >
            Result columns
          </CollapsibleSectionTitle>
        </div>

        {#if showColumns}
          <div transition:slide={{ duration: LIST_SLIDE_DURATION }}>
            <ul class="flex flex-col">
              {#each fields as field (field.name)}
                <li class="column-row">
                  <DataTypeIcon
                    type={field.type?.code?.replace(/^CODE_/, "") ?? "UNKNOWN"}
                    suppressTooltip
                  />
                  <span class="truncate text-xs" title={field.name}>
                    {field.name}
                  </span>
                  <span class="text-fg-secondary text-[10px] ml-auto flex-none">
                    {field.type?.code?.replace(/^CODE_/, "") ?? ""}
                  </span>
                </li>
              {/each}
            </ul>
          </div>
        {/if}
      </div>
    {:else}
      <div class="px-4 py-24 italic text-fg-disabled text-center">
        Run a query to see schema
      </div>
    {/if}
  </div>
</Inspector>

<style lang="postcss">
  .column-row {
    @apply flex items-center gap-x-2 px-4 py-1;
  }

  .column-row:hover {
    @apply bg-popover-accent;
  }
</style>
