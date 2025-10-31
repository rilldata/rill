<script lang="ts">
  import ExploreIcon from "@rilldata/web-common/components/icons/ExploreIcon.svelte";
  import * as Tooltip from "@rilldata/web-common/components/tooltip-v2/index.ts";
  import {
    ContextTypeData,
    FILTER_CONTEXT_TYPES,
  } from "@rilldata/web-common/features/chat/core/context/context-type-data.ts";
  import { ConversationContext } from "@rilldata/web-common/features/chat/core/context/context.ts";
  import { getAttrs, builderActions } from "bits-ui";
  import { ConversationContextType } from "web-common/src/features/chat/core/context/context-type-data.ts";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store.ts";

  export let context: ConversationContext;

  $: ({ instanceId } = $runtime);

  $: contextRecord = context.record;

  $: dashboardContextEntry = $contextRecord[ConversationContextType.Explore];
  $: filtersContextCount = Object.keys($contextRecord).filter((t) =>
    FILTER_CONTEXT_TYPES.includes(t as ConversationContextType),
  ).length;

  const timeRangeFormatter =
    ContextTypeData[ConversationContextType.TimeRange].formatter;
  $: formattedTimeRange = timeRangeFormatter(
    $contextRecord[ConversationContextType.TimeRange],
    $contextRecord,
    instanceId,
  );

  const whereFilterFormatter =
    ContextTypeData[ConversationContextType.Where].formatter;
  $: formattedWhereFilter = whereFilterFormatter(
    $contextRecord[ConversationContextType.Where],
    $contextRecord,
    instanceId,
  );
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
          on:click
        >
          @ {filtersContextCount} filter(s)
        </button>
      </Tooltip.Trigger>

      <Tooltip.Content
        class="flex flex-col gap-y-2 rounded-md border bg-popover text-popover-foreground shadow-md text-sm"
      >
        <div>{$formattedTimeRange}</div>
        <div>{$formattedWhereFilter}</div>
      </Tooltip.Content>
    </Tooltip.Root>
  {/if}
</div>
