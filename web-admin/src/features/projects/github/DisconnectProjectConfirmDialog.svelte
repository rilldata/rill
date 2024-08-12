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
  import { Button } from "@rilldata/web-common/components/button/index.js";

  export let open = false;
  export let loading = false;
  export let onDisconnect: () => void;
</script>

<AlertDialog bind:open>
  <AlertDialogTrigger asChild>
    <div class="hidden"></div>
  </AlertDialogTrigger>
  <AlertDialogContent>
    <AlertDialogHeader>
      <AlertDialogTitle>Disconnect from GitHub?</AlertDialogTitle>
      <AlertDialogDescription>
        <div class="mt-1">
          This project will be disconnected from GitHub and no longer under
          version control. Are you sure want to disconnect?
          <a
            href="https://docs.rilldata.com/deploy/existing-project/github-101"
            target="_blank"
            class="text-primary-600"
          >
            Learn more ->
          </a>
        </div>
      </AlertDialogDescription>
    </AlertDialogHeader>
    <AlertDialogFooter>
      <Button type="secondary" on:click={() => (open = false)}>Cancel</Button>
      <Button type="primary" on:click={onDisconnect} {loading}>
        Yes, disconnect
      </Button>
    </AlertDialogFooter>
  </AlertDialogContent>
</AlertDialog>
