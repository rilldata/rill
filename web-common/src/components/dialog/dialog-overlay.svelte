<script lang="ts">
  import { cn } from "@rilldata/web-common/lib/shadcn";
  import { Dialog as DialogPrimitive } from "bits-ui";

  type $$Props = DialogPrimitive.OverlayProps;

  let className: $$Props["class"] = undefined;
  export { className as class };
</script>

<DialogPrimitive.Overlay
  class={cn(
    "fixed inset-0 bg-neutral-800 dark:bg-neutral-100 transition-opacity opacity-60 z-50",
    className,
  )}
  {...$$restProps}
/>
