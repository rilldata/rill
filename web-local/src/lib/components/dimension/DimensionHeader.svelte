<script lang="ts">
  import { slideRight } from "../../transitions";
  import { EntityStatus } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/EntityStateService";

  import Back from "../icons/Back.svelte";
  import { metricsExplorerStore } from "../../application-state-stores/explorer-stores";
  import Spinner from "../Spinner.svelte";

  export let metricsDefId: string;
  export let isFetching: boolean;

  const goBackToLeaderboard = () => {
    metricsExplorerStore.setMetricDimensionId(metricsDefId, null);
  };
</script>

<div class="grid grid-auto-cols justify-start grid-flow-col items-end p-1 pb-3">
  <button
    on:click={() => goBackToLeaderboard()}
    class="flex flex-row items-center mb-4"
    style:grid-column-gap=".4rem"
  >
    {#if isFetching}
      <div transition:slideRight|local={{ leftOffset: 8 }}>
        <Spinner size="16px" status={EntityStatus.Running} />
      </div>
    {:else}
      <Back size="16px" />
      <span> All Dimensions </span>
    {/if}
  </button>
</div>
