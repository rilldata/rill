<script lang="ts">
  import { Dialog as DialogPrimitive } from "bits-ui";

  // svelte-ignore custom_element_props_identifier
  let { ...restProps }: DialogPrimitive.CloseProps = $props();
</script>

<DialogPrimitive.Close {...restProps} />
