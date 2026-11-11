<script lang="ts">
  import * as AlertDialog from "@rilldata/web-common/components/alert-dialog";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import type { BlockKind } from "./stores/block-edit";

  // Asks to confirm deleting a block that has content. It's open while mounted.
  let {
    kind,
    onConfirm,
    onCancel,
  }: {
    kind: BlockKind;
    onConfirm: () => void;
    onCancel: () => void;
  } = $props();

  const title = $derived(
    kind === "tab-group"
      ? m.canvas_delete_tab_group_title()
      : m.canvas_delete_container_title(),
  );
  const description = $derived(
    kind === "tab-group"
      ? m.canvas_delete_tab_group_description()
      : m.canvas_delete_container_description(),
  );
</script>

<AlertDialog.Confirmation
  open
  onOpenChange={(open) => {
    if (!open) onCancel();
  }}
  {title}
  {description}
  confirmLabel={m.common_delete()}
  confirmType="destructive"
  {onConfirm}
/>
