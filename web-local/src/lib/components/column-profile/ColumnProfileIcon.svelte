<script lang="ts">
  import { DataTypeIcon } from "@rilldata/web-common/components/data-types";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import { LIST_SLIDE_DURATION } from "@rilldata/web-common/layout/config";
  import { fade } from "svelte/transition";

  export let isFetching = false;
  export let type: string;
</script>

<div style:width="16px" style:height="16px">
  <div class="relative" style:width="16px" style:height="16px">
    {#key isFetching}
      <div
        class="absolute m-auto grid place-items-center"
        transition:fade|local={{ duration: LIST_SLIDE_DURATION }}
        style:width="16px"
        style:height="16px"
      >
        {#if isFetching}
          <Spinner status={EntityStatus.Running} bg="rgba(0,0,0,.2)" />
        {:else}
          <DataTypeIcon {type} />
        {/if}
      </div>
    {/key}
  </div>
</div>
