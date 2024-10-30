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
  import Champagne from "@rilldata/web-common/components/icons/Champagne.svelte";

  export let open: boolean;
</script>

<AlertDialog bind:open>
  <AlertDialogTrigger asChild>
    <div class="hidden"></div>
  </AlertDialogTrigger>
  <AlertDialogContent class="flex flex-row gap-x-2 min-w-[600px]">
    <Champagne size="150px" className="min-w-[150px]" />
    <div class="flex flex-col gap-x-2">
      <AlertDialogHeader>
        <AlertDialogTitle>Welcome to Rill Cloud</AlertDialogTitle>
        <AlertDialogDescription>
          Congrats on starting your <b>Team</b> plan. To get the most out of
          your plan,
          <a
            href="https://docs.rilldata.com/"
            target="_blank"
            class="text-primary-600 font-medium">refer to our docs</a
          >
        </AlertDialogDescription>
      </AlertDialogHeader>
      <div class="grow"></div>
      <AlertDialogFooter class="mt-3">
        <Button type="primary" on:click={() => (open = false)}>Got it</Button>
      </AlertDialogFooter>
    </div>
  </AlertDialogContent>
</AlertDialog>
