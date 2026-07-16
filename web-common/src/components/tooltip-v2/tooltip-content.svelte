<script lang="ts">
  import { Tooltip as TooltipPrimitive } from "bits-ui";
  import { getThemeBoundaryClass } from "@rilldata/web-common/features/themes/theme-boundary";
  import { cn } from "@rilldata/web-common/lib/shadcn";

  type $$Props = TooltipPrimitive.ContentProps;

  let className: $$Props["class"] = undefined;
  export let sideOffset: $$Props["sideOffset"] = 4;
  export { className as class };

  const themeBoundaryClass = getThemeBoundaryClass();
</script>

<TooltipPrimitive.Portal>
  <TooltipPrimitive.Content
    {sideOffset}
    class={cn(
      themeBoundaryClass,
      "bg-tooltip text-fg-inverse border shadow-md z-50 overflow-hidden rounded-md px-3 py-1.5 text-xs max-w-[min(24rem,calc(100vw-16px))] break-words",
      className,
    )}
    {...$$restProps}
  >
    <slot />
  </TooltipPrimitive.Content>
</TooltipPrimitive.Portal>
