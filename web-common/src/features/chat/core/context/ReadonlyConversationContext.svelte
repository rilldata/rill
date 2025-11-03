<script lang="ts">
  import ExploreIcon from "@rilldata/web-common/components/icons/ExploreIcon.svelte";
  import * as Tooltip from "@rilldata/web-common/components/tooltip-v2/index.ts";
  import { FILTER_CONTEXT_TYPES } from "@rilldata/web-common/features/chat/core/context/context-type-data.ts";
  import { ConversationContext } from "@rilldata/web-common/features/chat/core/context/context.ts";
  import {
    createTimeRangeFormatter,
    createWhereFiltersFormatter,
  } from "@rilldata/web-common/features/chat/core/context/formatters.ts";
  import { getAttrs, builderActions } from "bits-ui";
  import { ConversationContextType } from "web-common/src/features/chat/core/context/context-type-data.ts";

  export let context: ConversationContext;

  $: contextRecord = context.record;

  $: dashboardContextEntry = $contextRecord[ConversationContextType.Explore];
  $: filtersContextCount = Object.keys($contextRecord).filter((t) =>
    FILTER_CONTEXT_TYPES.includes(t as ConversationContextType),
  ).length;

  $: formattedTimeRange = createTimeRangeFormatter(context);

  $: formattedWhereFilters = createWhereFiltersFormatter(context);
</script>

<div class="flex flex-row items-center gap-2">
  {#if dashboardContextEntry}
    <div class="flex flex-row items-center gap-1 text-white">
      <ExploreIcon size="16px" />
      <span>{dashboardContextEntry}</span>
    </div>
  {/if}

  {#if filtersContextCount}
    <Tooltip.Root>
      <Tooltip.Trigger asChild let:builder>
        <button
          {...getAttrs([builder])}
          use:builderActions={{ builders: [builder] }}
          class="text-white"
        >
          @ {filtersContextCount} filter(s)
        </button>
      </Tooltip.Trigger>

      <Tooltip.Content
        class="flex flex-col gap-y-2 max-w-[250px] rounded-md border bg-popover text-popover-foreground shadow-md text-sm"
      >
        <div class="h-5 overflow-hidden whitespace-nowrap text-ellipsis">
          {$formattedTimeRange}
        </div>
        {#each $formattedWhereFilters as filter, i (i)}
          <div class="h-5 overflow-hidden whitespace-nowrap text-ellipsis">
            {filter}
          </div>
        {/each}
      </Tooltip.Content>
    </Tooltip.Root>
  {/if}
</div>
