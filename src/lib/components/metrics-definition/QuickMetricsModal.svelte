<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import { Dialog } from "../modal-new";

  const dispatch = createEventDispatcher();
</script>

<Dialog showCancel on:cancel on:submit={() => dispatch("replace-metrics")}>
  <svelte:fragment slot="title">Replace current metrics?</svelte:fragment>
  <svelte:fragment slot="body"
    >Continuing will replace all existing measures and dimensions.
  </svelte:fragment>
  <svelte:fragment slot="submit-body">Replace metrics</svelte:fragment>
</Dialog>
