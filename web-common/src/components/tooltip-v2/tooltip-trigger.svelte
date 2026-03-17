<script lang="ts">
  import { Tooltip as TooltipPrimitive } from "bits-ui";

  // svelte-ignore custom_element_props_identifier
  let { ...restProps }: TooltipPrimitive.TriggerProps = $props();
</script>

<TooltipPrimitive.Trigger {...restProps} />
