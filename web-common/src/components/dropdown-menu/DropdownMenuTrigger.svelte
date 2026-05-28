<script lang="ts">
  import { DropdownMenu as DropdownMenuPrimitive } from "bits-ui";

  // svelte-ignore custom_element_props_identifier
  let { ...restProps }: DropdownMenuPrimitive.TriggerProps = $props();
</script>

<DropdownMenuPrimitive.Trigger {...restProps} />
