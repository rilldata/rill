<script lang="ts">
  import {
    DialogContent,
    DialogTrigger,
  } from "@rilldata/web-common/components/dialog-v2";
  import GuardedDialog from "@rilldata/web-common/components/dialog-v2/GuardedDialog.svelte";
  import EditAlertForm from "@rilldata/web-common/features/alerts/EditAlertForm.svelte";
  import type { V1AlertSpec } from "@rilldata/web-common/runtime-client";
  import Button from "web-common/src/components/button/Button.svelte";

  export let alertSpec: V1AlertSpec;
  export let metricsViewName: string;
</script>

<GuardedDialog
  title="Close without saving?"
  description="You haven’t saved changes to this alert yet, so closing this window will lose your work."
  confirmLabel="Close"
  cancelLabel="Keep editing"
  let:onCancel
  let:onClose
>
  <DialogTrigger asChild let:builder>
    <Button type="secondary" builders={[builder]}>Edit</Button>
  </DialogTrigger>
  <DialogContent class="p-0 m-0 w-[602px] max-w-fit">
    <EditAlertForm
      {alertSpec}
      on:cancel={onCancel}
      on:close={onClose}
      {metricsViewName}
    />
  </DialogContent>
</GuardedDialog>
