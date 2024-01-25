<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import PivotPanel from "@rilldata/web-common/components/icons/PivotPanel.svelte";
  import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import { getStateManagers } from "../state-managers/state-managers";

  export let showPanels = true;
  export let isFetching = false;

  const stateManagers = getStateManagers();
  const { metricsViewName } = stateManagers;

  $: nestedMode = false;

  function toggleNestedMode() {
    console.log("toggleNestedMode");
  }
</script>

<div class="flex items-center gap-x-4 p-2 px-4">
  <Button
    square
    type="secondary"
    selected={showPanels}
    on:click={(e) => {
      showPanels = !showPanels;
      e.detail.currentTarget.blur();
    }}
  >
    <PivotPanel size="18px" />
  </Button>

  <Button
    compact
    type="text"
    on:click={() => {
      metricsExplorerStore.setPivotExpanded($metricsViewName, {});
    }}
  >
    Expand Visible
  </Button>

  <Button
    compact
    type="text"
    on:click={() => {
      metricsExplorerStore.setPivotExpanded($metricsViewName, {});
    }}
  >
    Collapse All
  </Button>
  {#if isFetching}
    <Spinner size="18px" status={EntityStatus.Running} />
  {/if}
</div>
