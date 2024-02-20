<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import { Button } from "../../../components/button";
  import CliCommandDisplay from "../../../components/commands/CLICommandDisplay.svelte";
  import Dialog from "../../../components/dialog/Dialog.svelte";

  export let open: boolean;

  const dispatch = createEventDispatcher();

  function close() {
    dispatch("close");
  }
</script>

<Dialog titleMarginBottomOverride="mb-4" on:close {open}>
  <svelte:fragment slot="title">
    Deploy your project to Rill Cloud
  </svelte:fragment>

  <div class="flex flex-col items-center" slot="body">
    <div class="text-left text-sm text-gray-500 w-full">
      Run this command from your project directory. <a
        href="https://docs.rilldata.com/deploy/existing-project"
        target="_blank"
        rel="noreferrer">See docs</a
      >
    </div>
    <div class="pt-4 pb-2">
      <CliCommandDisplay command="rill deploy" />
    </div>
  </div>

  <svelte:fragment slot="footer">
    <div class="flex">
      <div class="grow" />
      <Button type="secondary" on:click={close}>Close</Button>
    </div>
  </svelte:fragment>
</Dialog>
