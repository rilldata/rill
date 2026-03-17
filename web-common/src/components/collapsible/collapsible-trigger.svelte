<script lang="ts">
  import { Collapsible as CollapsiblePrimitive } from "bits-ui";

  // svelte-ignore custom_element_props_identifier
  let { ...restProps }: CollapsiblePrimitive.TriggerProps = $props();
</script>

<CollapsiblePrimitive.Trigger {...restProps} />
