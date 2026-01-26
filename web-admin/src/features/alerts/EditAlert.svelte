<script lang="ts">
  import {
    DialogContent,
    DialogTrigger,
  } from "@rilldata/web-common/components/dialog";
  import GuardedDialog from "@rilldata/web-common/components/dialog/GuardedDialog.svelte";
  import AlertForm from "@rilldata/web-common/features/alerts/AlertForm.svelte";
  import type { V1AlertSpec } from "@rilldata/web-common/runtime-client/gen/index.schemas";
  import Button from "web-common/src/components/button/Button.svelte";

  export let alertSpec: V1AlertSpec;
  export let disabled: boolean;
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
    <Button type="secondary" builders={[builder]} {disabled}>Edit</Button>
  </DialogTrigger>
  <DialogContent class="p-0 m-0 w-[802px] max-w-fit" noClose>
    <AlertForm props={{ mode: "edit", alertSpec }} {onCancel} {onClose} />
  </DialogContent>
</GuardedDialog>
