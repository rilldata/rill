<script lang="ts">
  import { Dialog } from "@rilldata/web-common/components/modal";
  import { createEventDispatcher } from "svelte";

  const dispatch = createEventDispatcher();
</script>

<Dialog
  showCancel
  on:cancel
  on:click-outside
  on:primary-action={() => dispatch("replace-metrics")}
>
  <svelte:fragment slot="title">Replace current metrics?</svelte:fragment>
  <svelte:fragment slot="body"
    >Continuing will replace all existing measures and dimensions.
  </svelte:fragment>
  <svelte:fragment slot="primary-action-body">Replace metrics</svelte:fragment>
</Dialog>
