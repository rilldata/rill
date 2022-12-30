<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import { Dialog } from "@rilldata/web-common/components/modal/index";

  export let exportErrorMessage = "";
</script>

<Dialog>
  <svelte:fragment slot="title">Export error</svelte:fragment>
  <svelte:fragment slot="body">{exportErrorMessage}</svelte:fragment>
  <Button slot="footer" type="text">close</Button>
</Dialog>
