<script lang="ts">
  import { Tooltip as TooltipPrimitive } from "bits-ui";

  let { ...restProps }: TooltipPrimitive.TriggerProps = $props();
</script>

<TooltipPrimitive.Trigger {...restProps} />
