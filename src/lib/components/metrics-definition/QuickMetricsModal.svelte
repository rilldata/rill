<script lang="ts">
  import { createEventDispatcher } from "svelte";

  import { Button } from "../button";
  import { Dialog } from "../modal-new";

  const dispatch = createEventDispatcher();
</script>

<Dialog showCancel on:cancel>
  <svelte:fragment slot="title">Replace current metrics?</svelte:fragment>
  <svelte:fragment slot="body"
    >Continuing will replace all existing measures and dimensions.
  </svelte:fragment>
  <svelte:fragment slot="footer">
    <Button type="text" on:click={() => dispatch("cancel")}>Cancel</Button>
    <Button type="primary" on:click={() => dispatch("replace-metrics")}
      >Replace metrics</Button
    >
  </svelte:fragment>
</Dialog>
