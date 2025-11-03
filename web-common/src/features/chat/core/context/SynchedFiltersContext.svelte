<script lang="ts">
  import { beforeNavigate } from "$app/navigation";
  import { page } from "$app/stores";
  import Cancel from "@rilldata/web-common/components/icons/Cancel.svelte";
  import * as Tooltip from "@rilldata/web-common/components/tooltip-v2/index.ts";
  import { ConversationContextType } from "@rilldata/web-common/features/chat/core/context/context-type-data.ts";
  import {
    formatV1Expression,
    formatV1TimeRange,
  } from "@rilldata/web-common/features/chat/core/context/formatters.ts";
  import { Conversation } from "@rilldata/web-common/features/chat/core/conversation.ts";
  import { getDashboardResourceFromPage } from "@rilldata/web-common/features/dashboards/nav-utils.ts";
  import { useStableExploreState } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores.ts";
  import {
    copyFilterExpression,
    isExpressionEmpty,
  } from "@rilldata/web-common/features/dashboards/stores/filter-utils.ts";
  import { createStableTimeControlStoreFromName } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store.ts";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
  import { getAttrs, builderActions } from "bits-ui";
  import { writable } from "svelte/store";

  export let conversation: Conversation;

  $: context = conversation.context;
  $: contextRecord = context.record;

  $: pageDashboardResource = getDashboardResourceFromPage($page);
  $: exploreName = pageDashboardResource?.name ?? "";
  $: onExplorePage = pageDashboardResource?.kind === ResourceKind.Explore;

  const exploreNameStore = writable("");
  $: exploreNameStore.set(exploreName);
  const exploreState = useStableExploreState(exploreNameStore);
  const timeControlsStore =
    createStableTimeControlStoreFromName(exploreNameStore);

  $: shouldShowFilters = onExplorePage && $timeControlsStore?.ready;

  $: timeRange = {
    start: $timeControlsStore?.timeStart,
    end: $timeControlsStore?.timeEnd,
  };
  $: timeRangeContext = $contextRecord[ConversationContextType.TimeRange];
  $: formattedTimeRange = formatV1TimeRange(timeRange);

  $: whereFilter = $exploreState?.whereFilter;
  $: filterIsAvailable = !isExpressionEmpty(whereFilter);
  $: availableFilters = 1 + (filterIsAvailable ? 1 : 0);
  $: formattedWhereFilters = formatV1Expression(whereFilter);

  // Where filters is only set if not empty. But time range should always be there.
  // We do not support non-timestamp explore right now.
  $: filtersActive = !!timeRangeContext;

  $: if (exploreName) {
    // Always set explore name
    context.set(ConversationContextType.Explore, exploreName);
  }
  $: if (filtersActive) {
    // Keep the values in sync
    context.set(ConversationContextType.TimeRange, timeRange);
    if (filterIsAvailable) {
      context.set(
        ConversationContextType.Where,
        copyFilterExpression(whereFilter),
      );
    } else {
      context.delete(ConversationContextType.Where);
    }
  }

  let open = false;

  function setFilters() {
    context.set(ConversationContextType.TimeRange, timeRange);
    if (filterIsAvailable) {
      context.set(ConversationContextType.Where, whereFilter);
    }
    open = true;
  }

  function clearFilters(e) {
    e.stopPropagation();
    context.delete(ConversationContextType.TimeRange);
    context.delete(ConversationContextType.Where);
  }

  beforeNavigate(({ to }) => {
    const toResource = to ? getDashboardResourceFromPage(to) : undefined;
    if (!toResource) {
      context.clear();
    }
  });
</script>

<svelte:window on:click={() => (open = false)} />

{#if shouldShowFilters}
  <Tooltip.Root bind:open>
    <Tooltip.Trigger asChild let:builder>
      <button
        {...getAttrs([builder])}
        use:builderActions={{ builders: [builder] }}
        class="flex flex-row items-center gap-1 w-fit m-1 py-0.5 px-1 border border-input rounded-md"
        class:bg-primary-50={filtersActive}
        on:click={setFilters}
        type="button"
      >
        <span
          class:text-primary={filtersActive}
          class:text-muted-foreground={!filtersActive}
        >
          @ {availableFilters} filter(s)
        </span>
        {#if filtersActive}
          <button on:click={clearFilters}>
            <Cancel className="text-primary-600" />
          </button>
        {/if}
      </button>
    </Tooltip.Trigger>

    <Tooltip.Content
      class="flex flex-col gap-y-2 max-w-[250px] rounded-md border bg-popover text-popover-foreground shadow-md text-sm"
    >
      <div class="h-5 overflow-hidden whitespace-nowrap text-ellipsis">
        {formattedTimeRange}
      </div>
      {#each formattedWhereFilters as filter, i (i)}
        <div class="h-5 overflow-hidden whitespace-nowrap text-ellipsis">
          {filter}
        </div>
      {/each}
    </Tooltip.Content>
  </Tooltip.Root>
{/if}
