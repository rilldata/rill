<script lang="ts">
  import Back from "@rilldata/web-common/components/icons/Back.svelte";
  import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import { Button } from "../../../components/button";

  export let exploreName: string;
  export let isFetching = false;

  function goBackToExplore() {
    metricsExplorerStore.setExpandedMeasureName(exploreName, undefined);
  }
</script>

<button class="flex items-center" on:click={() => goBackToExplore()}>
  {#if isFetching}
    <Spinner size="16px" status={EntityStatus.Running} />
  {:else}
    <Button type="link" forcedStyle="padding: 0; gap: 0px;">
      <Back size="16px" />
      <span>All measures</span>
    </Button>
  {/if}
</button>
