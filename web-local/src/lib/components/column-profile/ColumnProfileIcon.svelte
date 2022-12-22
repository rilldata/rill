<script lang="ts">
  import { DataTypeIcon } from "@rilldata/web-common/components/data-types";
  import { fade } from "svelte/transition";
  import { LIST_SLIDE_DURATION } from "../../application-config";
  import { EntityStatus } from "../../temp/entity";
  import Spinner from "../Spinner.svelte";

  export let isFetching = false;
  export let type: string;
</script>

<div style:width="16px" style:height="16px">
  <div class="relative" style:width="16px" style:height="16px">
    {#key isFetching}
      <div
        class="absolute m-auto grid place-items-center"
        transition:fade={{ duration: LIST_SLIDE_DURATION }}
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
