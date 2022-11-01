<script lang="ts">
  import Button from "../../button/Button.svelte";
  import { Dialog } from "../../modal/index";

  export let exportErrorMessage = "";
</script>

<Dialog>
  <svelte:fragment slot="title">Export error</svelte:fragment>
  <svelte:fragment slot="body">{exportErrorMessage}</svelte:fragment>
  <Button slot="footer" type="text">close</Button>
</Dialog>
