<script lang="ts">
  import { Tooltip as TooltipPrimitive } from "bits-ui";
  import { cn } from "@rilldata/web-common/lib/shadcn";

  type $$Props = TooltipPrimitive.ContentProps;

  let className: $$Props["class"] = undefined;
  export let sideOffset: $$Props["sideOffset"] = 4;
  export { className as class };
</script>

<TooltipPrimitive.Portal>
  <TooltipPrimitive.Content
    {sideOffset}
    class={cn(
      "bg-tooltip text-fg-inverse border shadow-md z-50 overflow-hidden rounded-md px-3 py-1.5 text-xs",
      className,
    )}
    {...$$restProps}
  >
    <slot />
  </TooltipPrimitive.Content>
</TooltipPrimitive.Portal>
