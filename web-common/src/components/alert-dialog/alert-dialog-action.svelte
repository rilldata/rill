<script lang="ts">
  import { AlertDialog as AlertDialogPrimitive } from "bits-ui";
  import { buttonVariants } from "./button-variants";
  import { cn } from "@rilldata/web-common/lib/shadcn";

  let { class: className, ...restProps }: AlertDialogPrimitive.ActionProps =
    $props();
</script>

<AlertDialogPrimitive.Action
  class={cn(buttonVariants(), className)}
  {...restProps}
/>
