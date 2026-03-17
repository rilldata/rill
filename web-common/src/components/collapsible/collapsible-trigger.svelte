<script lang="ts">
  import { Collapsible as CollapsiblePrimitive } from "bits-ui";

  let { ...restProps }: CollapsiblePrimitive.TriggerProps = $props();
</script>

<CollapsiblePrimitive.Trigger {...restProps} />
