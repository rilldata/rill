<script lang="ts">
  import { EntityStatus } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/EntityStateService";

  import FilterInclude from "@rilldata/web-local/lib/components/icons/FilterInclude.svelte";
  import FilterRemove from "@rilldata/web-local/lib/components/icons/FilterRemove.svelte";

  import Spinner from "@rilldata/web-local/lib/components/Spinner.svelte";
  import Shortcut from "@rilldata/web-local/lib/components/tooltip/Shortcut.svelte";
  import Tooltip from "@rilldata/web-local/lib/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-local/lib/components/tooltip/TooltipContent.svelte";
  import TooltipShortcutContainer from "@rilldata/web-local/lib/components/tooltip/TooltipShortcutContainer.svelte";
  import TooltipTitle from "@rilldata/web-local/lib/components/tooltip/TooltipTitle.svelte";

  export let displayName: string;
  export let isFetching: boolean;
  export let toggleFilterExcludeMode: () => void;
  export let dimensionDescription: string;
  export let hovered: boolean;
  export let filterKey: "exclude" | "include";
  export let filterExcludeMode: boolean;

  let otherFilterKey: "exclude" | "include";
  $: otherFilterKey = filterKey === "include" ? "exclude" : "include";
</script>

<div class="flex flex-row  items-center">
  <div style:width="22px" style:height="22px" class="grid place-items-center">
    {#if isFetching}
      <Spinner size="16px" status={EntityStatus.Running} />
    {/if}
  </div>

  <button
    style:height="32px"
    style:flex="1"
    style:grid-template-columns="auto max-content"
    class="
        pr-2
        grid justify-between items-center
        w-full
        border-b
        border-gray-200
        rounded-t
        bg-white
        text-gray-600
        font-semibold
    "
    on:click
  >
    <div>
      <Tooltip location="top" distance={16}>
        <div class="pl-2">{displayName}</div>
        <TooltipContent slot="tooltip-content">
          <TooltipTitle>
            <svelte:fragment slot="name">
              {displayName}
            </svelte:fragment>
            <svelte:fragment slot="description" />
          </TooltipTitle>
          <TooltipShortcutContainer>
            <div>
              {#if dimensionDescription}
                {dimensionDescription}
              {:else}
                the leaderboard metrics for {displayName}
              {/if}
            </div>
            <Shortcut />
            <div>Expand leaderboard</div>
            <Shortcut>Click</Shortcut>
          </TooltipShortcutContainer>
        </TooltipContent>
      </Tooltip>
    </div>
    <div>
      {#if hovered}
        <Tooltip location="top" distance={16}>
          <div on:click|stopPropagation={toggleFilterExcludeMode}>
            {#if filterExcludeMode}<FilterRemove
                size="16px"
              />{:else}<FilterInclude size="16px" />{/if}
          </div>
          <TooltipContent slot="tooltip-content">
            <TooltipTitle>
              <svelte:fragment slot="name">
                Output {filterKey}s selected values
              </svelte:fragment>
            </TooltipTitle>
            <TooltipShortcutContainer>
              <div>toggle to {otherFilterKey} values</div>
              <Shortcut>Click</Shortcut>
            </TooltipShortcutContainer>
          </TooltipContent>
        </Tooltip>
      {/if}
    </div>
  </button>
</div>
