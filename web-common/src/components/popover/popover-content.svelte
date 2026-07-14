<script lang="ts">
  import { cn } from "@rilldata/web-common/lib/shadcn";
  import { Popover as PopoverPrimitive } from "bits-ui";

  type $$Props = PopoverPrimitive.ContentProps & {
    overflowY?: string;
    overflowX?: string;
    minHeight?: string;
    padding?: string;
  };

  let className: $$Props["class"] = undefined;

  export let align: $$Props["align"] = "center";
  export let sideOffset: $$Props["sideOffset"] = 4;
  export let avoidCollisions: $$Props["avoidCollisions"] = true;
  export let collisionPadding: $$Props["collisionPadding"] = 8;
  export let collisionBoundary: $$Props["collisionBoundary"] = undefined;
  export let strategy: $$Props["strategy"] = undefined;
  export let overflowY: string = "auto";
  export let overflowX: string = "auto";
  export let minHeight: string = "0px";
  export let padding: string = "4";
  export { className as class };
</script>

<PopoverPrimitive.Portal>
  <PopoverPrimitive.Content
    {align}
    {sideOffset}
    {avoidCollisions}
    {collisionPadding}
    {collisionBoundary}
    {strategy}
    {...$$restProps}
    class={cn(
      "z-50 w-72 rounded-md border bg-popover text-popover-foreground shadow-md outline-none",
      overflowY && `overflow-y-${overflowY}`,
      overflowX && `overflow-x-${overflowX}`,
      minHeight && `min-h-[${minHeight}]`,
      padding && `p-${padding}`,
      className,
    )}
  >
    <slot />
  </PopoverPrimitive.Content>
</PopoverPrimitive.Portal>
