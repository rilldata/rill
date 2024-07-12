<script lang="ts">
  import {
    AlertDialog,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
    AlertDialogTrigger,
  } from "@rilldata/web-common/components/alert-dialog";
  import { Button } from "@rilldata/web-common/components/button";
  import Github from "@rilldata/web-common/components/icons/Github.svelte";

  export let open = false;
  export let onContinue: () => void;
</script>

<AlertDialog bind:open>
  <AlertDialogTrigger asChild>
    <div class="hidden"></div>
  </AlertDialogTrigger>
  <AlertDialogContent>
    <div class="flex flex-row gap-x-2">
      <Github size="28px" />
      <div class="flex flex-col">
        <AlertDialogHeader>
          <AlertDialogTitle>
            Connect this project to a Github repo
          </AlertDialogTitle>
          <AlertDialogDescription>
            Before continuing, you need to make sure your Rill project has been
            <!-- TODO: docs link -->
            pushed to a repo on Github.
            <a href="https://docs.rilldata.com">Learn how</a>
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter class="mt-5">
          <Button type="secondary" on:click={() => (open = false)}>
            Cancel
          </Button>
          <Button type="primary" on:click={onContinue}>Continue</Button>
        </AlertDialogFooter>
      </div>
    </div>
  </AlertDialogContent>
</AlertDialog>
