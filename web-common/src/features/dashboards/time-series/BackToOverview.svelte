<script lang="ts">
  import Back from "@rilldata/web-common/components/icons/Back.svelte";
  import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/dashboard-stores";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";

  export let metricViewName;
  export let isFetching = false;

  function goBackToOverview() {
    metricsExplorerStore.setExpandedMeasureName(metricViewName, undefined);
  }
</script>

<button
  class="flex flex-row"
  on:click={() => goBackToOverview()}
  style:grid-column-gap=".4rem"
>
  {#if isFetching}
    <div>
      <Spinner size="16px" status={EntityStatus.Running} />
    </div>
  {:else}
    <span class="ui-copy-icon">
      <Back size="16px" />
    </span>
    <span> Overview </span>
  {/if}
</button>
