<script lang="ts">
  import { DataTypeIcon } from "@rilldata/web-common/components/data-types";
  import Spinner from "@rilldata/web-common/features/temp/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/lib/entity";
  import { fade } from "svelte/transition";
  import { LIST_SLIDE_DURATION } from "../../application-config";

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
        <svelte:component
          this={isFetching ? Spinner : DataTypeIcon}
          status={EntityStatus.Running}
          {type}
          bg="rgba(0,0,0,.2)"
        />
      </div>
    {/key}
  </div>
</div>
