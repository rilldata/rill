<script lang="ts">
  import { slideRight } from "../../transitions";
  import { EntityStatus } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/EntityStateService";

  import { Button } from "../button";

  import Tooltip from "../tooltip/Tooltip.svelte";
  import TooltipContent from "../tooltip/TooltipContent.svelte";
  import TooltipTitle from "../tooltip/TooltipTitle.svelte";
  import TooltipShortcutContainer from "../tooltip/TooltipShortcutContainer.svelte";
  import Shortcut from "../tooltip/Shortcut.svelte";

  import Back from "../icons/Back.svelte";
  import { metricsExplorerStore } from "../../application-state-stores/explorer-stores";
  import Spinner from "../Spinner.svelte";
  import Spacer from "../icons/Spacer.svelte";

  export let metricsDefId: string;
  export let dimensionId: string;
  export let isFetching: boolean;
  export let excludeMode = false;

  $: filterKey = excludeMode ? "exclude" : "include";
  $: otherFilterKey = excludeMode ? "include" : "exclude";

  const goBackToLeaderboard = () => {
    metricsExplorerStore.setMetricDimensionId(metricsDefId, null);
  };
  function toggleFilterMode() {
    metricsExplorerStore.toggleFilterMode(metricsDefId, dimensionId);
  }
</script>

<div
  class="grid justify-start items-center pb-3"
  style:grid-template-columns="24px calc(100% - 24px)"
>
  <div transition:slideRight|local={{ leftOffset: 8 }}>
    {#if isFetching}
      <Spinner size="16px" status={EntityStatus.Running} />
    {:else}
      <Spacer size="16px" />
    {/if}
  </div>
  <div
    class="grid justify-between items-center"
    style:grid-template-columns="auto max-content"
  >
    <Button type="secondary" on:click={goBackToLeaderboard} compact>
      <Back size="16px" />All Dimensions
    </Button>

    <Tooltip location="left" distance={16}>
      <Button type="secondary" on:click={toggleFilterMode} compact>
        {#if excludeMode}exclude{:else}include{/if}
      </Button>
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
  </div>
</div>
