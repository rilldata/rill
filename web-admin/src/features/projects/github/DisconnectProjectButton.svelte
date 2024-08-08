<script lang="ts">
  import {
    AlertDialog,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
    AlertDialogTrigger,
  } from "@rilldata/web-common/components/alert-dialog/index.js";
  import EditIcon from "@rilldata/web-common/components/icons/EditIcon.svelte";
  import Github from "@rilldata/web-common/components/icons/Github.svelte";
  import { Button } from "@rilldata/web-common/components/button/index.js";

  export let open = false;
  export let loading = false;
  export let showButton = false;
  export let onDisconnect: () => void;
</script>

<AlertDialog bind:open>
  <AlertDialogTrigger asChild let:builder>
    <Button builders={[builder]} type="ghost" compact>
      <div class="min-w-4">
        {#if showButton}
          <EditIcon size="16px" />
        {/if}
      </div>
    </Button>
  </AlertDialogTrigger>
  <AlertDialogContent>
    <AlertDialogHeader>
      <div class="flex flex-row gap-x-2 items-center">
        <Github size="40px" />
        <AlertDialogTitle>Disconnect GitHub</AlertDialogTitle>
      </div>

      <AlertDialogDescription>
        <div class="mt-1">
          Do you want to disconnect the project from the GitHub repo?
        </div>
      </AlertDialogDescription>
    </AlertDialogHeader>
    <AlertDialogFooter class="mt-3">
      <Button type="secondary" on:click={() => (open = false)}>Cancel</Button>
      <Button type="primary" on:click={onDisconnect} {loading}>
        Disconnect
      </Button>
    </AlertDialogFooter>
  </AlertDialogContent>
</AlertDialog>
