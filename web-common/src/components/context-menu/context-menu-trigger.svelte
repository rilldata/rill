<script lang="ts">
  import { ContextMenu as ContextMenuPrimitive } from "bits-ui";

  let { ...restProps }: ContextMenuPrimitive.TriggerProps = $props();
</script>

<ContextMenuPrimitive.Trigger {...restProps} />
