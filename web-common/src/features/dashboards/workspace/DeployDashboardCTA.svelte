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
  import DeployIcon from "@rilldata/web-common/components/icons/DeployIcon.svelte";
  import { deploy } from "@rilldata/web-common/features/project/deploy";
  import { createEventDispatcher } from "svelte";
  import { Button } from "../../../components/button";

  export let open: boolean;

  const dispatch = createEventDispatcher();

  function close() {
    dispatch("close");
  }

  function onDeploy() {
    void deploy();
  }
</script>

<AlertDialog bind:open>
  <AlertDialogTrigger asChild>
    <div class="hidden"></div>
  </AlertDialogTrigger>
  <AlertDialogContent>
    <AlertDialogHeader>
      <div class="flex flex-row">
        <DeployIcon size="150px" />
        <div class="flex flex-col">
          <AlertDialogTitle>Deploy this project for free</AlertDialogTitle>
          <AlertDialogDescription>
            You’re about to start a 30-day FREE trial of Rill Cloud, where you
            can set alerts, share dashboards, and more.
          </AlertDialogDescription>
        </div>
      </div>
    </AlertDialogHeader>
    <AlertDialogFooter>
      <Button type="secondary" on:click={close}>Back</Button>
      <Button type="primary" on:click={onDeploy}>Continue</Button>
    </AlertDialogFooter>
  </AlertDialogContent>
</AlertDialog>
