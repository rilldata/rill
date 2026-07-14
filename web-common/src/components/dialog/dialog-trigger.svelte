<script lang="ts">
  import { Dialog as DialogPrimitive } from "bits-ui";

  // svelte-ignore custom_element_props_identifier
  let { ...restProps }: DialogPrimitive.TriggerProps = $props();
</script>

<DialogPrimitive.Trigger {...restProps} />
