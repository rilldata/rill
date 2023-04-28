<script lang="ts">
  import Dialog from "@rilldata/web-common/components/modal/dialog/Dialog.svelte";
  import { projectShareStore } from "@rilldata/web-common/features/dashboards/dashboard-stores";

  function close() {
    projectShareStore.set(false);
  }
</script>

<Dialog on:cancel={close} size="md">
  <div class="" slot="body">
    <div class="text-center text-2xl p-6">
      Please use the CLI to deploy your project by running `rill deploy`
    </div>
  </div>
</Dialog>
