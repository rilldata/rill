<script lang="ts">
  import { AlertDialog as AlertDialogPrimitive } from "bits-ui";
  import { buttonVariants } from "./button-variants";
  import { cn } from "@rilldata/web-common/lib/shadcn";

  let { class: className, ...restProps }: AlertDialogPrimitive.CancelProps =
    $props();
</script>

<AlertDialogPrimitive.Cancel
  class={cn(buttonVariants({ variant: "outline" }), "mt-2 sm:mt-0", className)}
  {...restProps}
/>
