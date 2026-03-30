<script lang="ts">
  import * as Dialog from "@rilldata/web-common/components/dialog";
  import AddDataManager from "@rilldata/web-common/features/add-data/manager/AddDataManager.svelte";

  export let open: boolean = true;
  export let schema: string | undefined = undefined;
  export let connector: string | undefined = undefined;

  // Use a boolean to mount remount when the modal is re-opened.
  // It is used to enure there is no stale state.
  let showForm = false;

  $: if (open) showForm = true;

  function handleDialogClose() {
    open = false;
    showForm = false;
  }
</script>

<Dialog.Root
  bind:open
  onOpenChange={(newOpen) => {
    if (!newOpen) showForm = false;
  }}
>
  <Dialog.Content class="p-0 w-fit max-w-fit h-fit" noClose>
    {#if showForm}
      <AddDataManager
        config={{ importOnly: true }}
        initSchema={schema}
        initConnector={connector}
        onClose={handleDialogClose}
        onDone={handleDialogClose}
      />
    {/if}
  </Dialog.Content>
</Dialog.Root>
