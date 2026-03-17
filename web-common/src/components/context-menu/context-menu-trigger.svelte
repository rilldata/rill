<script lang="ts">
  import { ContextMenu as ContextMenuPrimitive } from "bits-ui";

  // svelte-ignore custom_element_props_identifier
  let { ...restProps }: ContextMenuPrimitive.TriggerProps = $props();
</script>

<ContextMenuPrimitive.Trigger {...restProps} />
