<script lang="ts">
  import { AlertDialog as AlertDialogPrimitive } from "bits-ui";

  let { ...restProps }: AlertDialogPrimitive.TriggerProps = $props();
</script>

<AlertDialogPrimitive.Trigger {...restProps} />
