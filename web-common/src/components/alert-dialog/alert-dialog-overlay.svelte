<script lang="ts">
  import { AlertDialog as AlertDialogPrimitive } from "bits-ui";
  import { cn } from "@rilldata/web-common/lib/shadcn";

  type $$Props = AlertDialogPrimitive.OverlayProps;

  let className: $$Props["class"] = undefined;
  export { className as class };
</script>

<AlertDialogPrimitive.Overlay
  class={cn(
    "fixed inset-0 z-50 bg-gray-400 transition-opacity opacity-40",
    className,
  )}
  {...$$restProps}
/>
