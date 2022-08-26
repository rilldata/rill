<script lang="ts">
  import Button from "../button/Button.svelte";
  import { Dialog } from "../modal";

  export let exportErrorMessage = "";
</script>

<Dialog>
  <svelte:fragment slot="title">Export error</svelte:fragment>
  <svelte:fragment slot="body">{exportErrorMessage}</svelte:fragment>
  <svelte:fragment slot="footer">
    <Button type="text">close</Button>
  </svelte:fragment>
</Dialog>
