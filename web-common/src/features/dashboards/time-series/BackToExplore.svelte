<script lang="ts">
  import Back from "@rilldata/web-common/components/icons/Back.svelte";
  import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";

  export let exploreName: string;
  export let isFetching = false;

  function goBackToExplore() {
    metricsExplorerStore.setExpandedMeasureName(exploreName, undefined);
  }
</script>

<button
  class="flex flex-row items-center gap-x-1"
  on:click={() => goBackToExplore()}
>
  {#if isFetching}
    <Spinner size="16px" status={EntityStatus.Running} />
  {:else}
    <span class="ui-copy-icon">
      <Back color="var(--color-primary-600)" size="16px" />
    </span>
    <span class="text-primary-600 font-medium">All measures</span>
  {/if}
</button>
