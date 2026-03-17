<script lang="ts">
  import { Popover as PopoverPrimitive } from "bits-ui";

  let { ...restProps }: PopoverPrimitive.TriggerProps = $props();
</script>

<PopoverPrimitive.Trigger {...restProps} />
