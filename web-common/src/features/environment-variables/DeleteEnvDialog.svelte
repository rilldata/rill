<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import {
    AlertDialog,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
    AlertDialogTrigger,
  } from "@rilldata/web-common/components/alert-dialog/";
  import { createEventDispatcher } from "svelte";

  export let open = false;
  export let keyName: string;

  const dispatch = createEventDispatcher<{
    confirm: { key: string };
  }>();

  function handleConfirm() {
    dispatch("confirm", { key: keyName });
    open = false;
  }

  function handleCancel() {
    open = false;
  }
</script>

<AlertDialog bind:open>
  <AlertDialogTrigger asChild>
    <div class="hidden"></div>
  </AlertDialogTrigger>
  <AlertDialogContent>
    <AlertDialogHeader>
      <AlertDialogTitle>Delete environment variable</AlertDialogTitle>
    </AlertDialogHeader>
    <AlertDialogDescription>
      Are you sure you want to delete the environment variable <strong
        >{keyName}</strong
      >?
    </AlertDialogDescription>
    <AlertDialogFooter>
      <Button type="plain" onClick={handleCancel}>Cancel</Button>
      <Button type="primary" status="error" onClick={handleConfirm}>Delete</Button>
    </AlertDialogFooter>
  </AlertDialogContent>
</AlertDialog>
