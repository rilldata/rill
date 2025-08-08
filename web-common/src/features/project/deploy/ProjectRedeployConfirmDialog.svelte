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
  import CloudIcon from "@rilldata/web-common/components/icons/CloudIcon.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";

  export let isLoading: boolean;
  export let onConfirm: () => void;

  let open = false;
  function onRedeploy() {
    open = false;
    onConfirm();
  }
</script>

<AlertDialog bind:open>
  <AlertDialogTrigger asChild let:builder>
    <Tooltip distance={8}>
      <Button loading={isLoading} type="secondary" builders={[builder]}>
        <CloudIcon size="16px" />
        Update
      </Button>
      <TooltipContent slot="tooltip-content">
        Push changes to Rill Cloud
      </TooltipContent>
    </Tooltip>
  </AlertDialogTrigger>
  <AlertDialogContent>
    <AlertDialogHeader>
      <AlertDialogTitle>Push updates to Rill Cloud?</AlertDialogTitle>
      <AlertDialogDescription>
        <div class="mt-1">
          Would you like to send local changes to the deployed version of this
          project?
        </div>
      </AlertDialogDescription>
    </AlertDialogHeader>
    <AlertDialogFooter>
      <Button
        type="plain"
        onClick={() => {
          open = false;
        }}
      >
        Cancel
      </Button>
      <Button type="primary" onClick={onRedeploy}>Yes, update</Button>
    </AlertDialogFooter>
  </AlertDialogContent>
</AlertDialog>
