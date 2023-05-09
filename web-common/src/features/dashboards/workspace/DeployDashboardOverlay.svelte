<script lang="ts">
  import Dialog from "@rilldata/web-common/components/modal/dialog/Dialog.svelte";
  import { projectShareStore } from "@rilldata/web-common/features/dashboards/dashboard-stores";
  import Button from "../../../components/button/Button.svelte";
  import CliCommandDisplay from "../../../components/commands/CLICommandDisplay.svelte";
  import DialogFooter from "../../../components/modal/dialog/DialogFooter.svelte";

  function close() {
    projectShareStore.set(false);
  }
</script>

<Dialog on:cancel={close} size="md">
  <svelte:fragment slot="title"
    >Deploy your project to Rill Cloud</svelte:fragment
  >
  <div class="flex flex-col items-center gap-y-4" slot="body">
    <div class="text-left text-sm text-gray-500 w-full">
      Run this command from your project directory. <a
        href="https://docs.rilldata.com/deploy/existing-project"
        target="_blank"
        rel="noreferrer">See docs</a
      >
    </div>
    <CliCommandDisplay command="rill deploy" />
  </div>
  <svelte:fragment slot="secondary-action-body">Close</svelte:fragment>
  <DialogFooter slot="footer">
    <Button type="secondary" on:click={close}>Close</Button>
  </DialogFooter>
</Dialog>
