<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import { Dialog } from "../modal";

  const dispatch = createEventDispatcher();
</script>

<Dialog
  showCancel
  on:cancel
  on:primary-action={() => dispatch("replace-metrics")}
>
  <svelte:fragment slot="title">Replace current metrics?</svelte:fragment>
  <svelte:fragment slot="body"
    >Continuing will replace all existing measures and dimensions.
  </svelte:fragment>
  <svelte:fragment slot="primary-action-body">Replace metrics</svelte:fragment>
</Dialog>
