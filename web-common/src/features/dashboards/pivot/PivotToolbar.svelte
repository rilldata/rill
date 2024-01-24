<script lang="ts">
  import { Switch } from "@rilldata/web-common/components/button";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import PivotPanel from "@rilldata/web-common/components/icons/PivotPanel.svelte";
  import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import { getStateManagers } from "../state-managers/state-managers";

  export let showPanels = true;

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
  <Switch checked={nestedMode} on:click={() => toggleNestedMode()}>
    Nested
  </Switch>
  <div>Expand All</div>

  <Button
    type="text"
    on:click={(e) => {
      metricsExplorerStore.setPivotExpanded($metricsViewName, {});
    }}
  >
    Collapse All
  </Button>
</div>
