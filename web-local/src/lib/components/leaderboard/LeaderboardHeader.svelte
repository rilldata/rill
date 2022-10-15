<script lang="ts">
  import { EntityStatus } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/EntityStateService";

  import Spinner from "@rilldata/web-local/lib/components/Spinner.svelte";
  import Shortcut from "@rilldata/web-local/lib/components/tooltip/Shortcut.svelte";
  import Tooltip from "@rilldata/web-local/lib/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-local/lib/components/tooltip/TooltipContent.svelte";
  import TooltipShortcutContainer from "@rilldata/web-local/lib/components/tooltip/TooltipShortcutContainer.svelte";
  import TooltipTitle from "@rilldata/web-local/lib/components/tooltip/TooltipTitle.svelte";
  import LeaderboardOptionsMenu from "./LeaderboardOptionsMenu.svelte";

  export let displayName: string;
  export let isFetching: boolean;
  export let dimensionDescription: string;
  export let hovered: boolean;

  export let filterExcludeMode: boolean;

  let optionsMenuActive = false;
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
    style="max-width: calc(100% - 22px);"
    class="
        pr-2
        grid justify-between items-center
        w-full
        border-b
        border-gray-200
        rounded-t
        surface
        ui-copy-label
        font-semibold
        truncate
    "
  >
    <div on:click>
      <Tooltip location="top" distance={16}>
        <div class="pl-2 truncate" style="max-width: calc(315px - 60px);">
          {displayName}
        </div>
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
      {#if hovered || optionsMenuActive}
        <LeaderboardOptionsMenu
          bind:optionsMenuActive
          on:toggle-filter-exclude-mode
          {filterExcludeMode}
        />
      {/if}
    </div>
  </button>
</div>
