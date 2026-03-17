<script lang="ts">
  import { AlertDialog as AlertDialogPrimitive } from "bits-ui";

  // svelte-ignore custom_element_props_identifier
  let { ...restProps }: AlertDialogPrimitive.TriggerProps = $props();
</script>

<AlertDialogPrimitive.Trigger {...restProps} />
