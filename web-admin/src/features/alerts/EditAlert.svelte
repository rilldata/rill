<script lang="ts">
  import AlertDialog from "@rilldata/web-common/components/alert-dialog/AlertDialog.svelte";
  import {
    Dialog,
    DialogContent,
    DialogTrigger,
  } from "@rilldata/web-common/components/dialog-v2";
  import EditAlertDialog from "@rilldata/web-common/features/alerts/EditAlertDialog.svelte";
  import type { V1AlertSpec } from "@rilldata/web-common/runtime-client";
  import Button from "web-common/src/components/button/Button.svelte";

  export let alertSpec: V1AlertSpec;
  export let metricsViewName: string;

  let showAlertDialog = false;
  let showCancelDialog = false;
  function onCancel() {
    showCancelDialog = true;
  }
</script>

<Dialog
  bind:open={showAlertDialog}
  closeOnEscape={false}
  onOutsideClick={(e) => {
    e.preventDefault();
    onCancel();
  }}
>
  <DialogTrigger asChild let:builder>
    <Button type="secondary" builders={[builder]}>Edit</Button>
  </DialogTrigger>
  <DialogContent class="p-0 m-0 w-[602px] max-w-fit">
    <EditAlertDialog {alertSpec} on:close={onCancel} {metricsViewName} />
  </DialogContent>
</Dialog>

<AlertDialog
  title="Close without saving?"
  description="You haven’t saved changes to this alert yet, so closing this window will lose your work."
  confirmLabel="Close"
  onConfirm={() => {
    showAlertDialog = false;
    showCancelDialog = false;
  }}
  cancelLabel="Keep editing"
  bind:open={showCancelDialog}
>
  <div slot="trigger"></div>
</AlertDialog>
