<script lang="ts">
  import {
    DialogContent,
    DialogTrigger,
  } from "@rilldata/web-common/components/dialog";
  import GuardedDialog from "@rilldata/web-common/components/dialog/GuardedDialog.svelte";
  import AlertFormDataWrapper from "@rilldata/web-common/features/alerts/AlertFormDataWrapper.svelte";
  import type { V1AlertSpec } from "@rilldata/web-common/runtime-client/gen/index.schemas";
  import * as m from "@rilldata/web-common/paraglide/messages.js";
  import Button from "web-common/src/components/button/Button.svelte";

  export let alertSpec: V1AlertSpec;
  export let disabled: boolean;
</script>

<GuardedDialog
  title={m.dialog_close_without_saving_title()}
  description={m.dialog_close_without_saving_alert_desc()}
  confirmLabel={m.dialog_close_without_saving_confirm()}
  cancelLabel={m.dialog_close_without_saving_cancel()}
  let:onCancel
  let:onClose
  let:preventClose
>
  <DialogTrigger>
    {#snippet child({ props })}
      <Button {...props} type="secondary" {disabled} label={m.alert_edit()}
        >{m.alert_form_update()}</Button
      >
    {/snippet}
  </DialogTrigger>
  <DialogContent
    class="p-0 m-0 w-[802px] max-w-fit"
    noClose
    onEscapeKeydown={preventClose}
    onInteractOutside={preventClose}
  >
    <AlertFormDataWrapper
      props={{ mode: "edit", alertSpec }}
      {onCancel}
      {onClose}
    />
  </DialogContent>
</GuardedDialog>
