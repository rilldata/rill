<script lang="ts">
  import { cn } from "@rilldata/web-common/lib/shadcn";
  import { DropdownMenu as DropdownMenuPrimitive } from "bits-ui";

  type $$Props = DropdownMenuPrimitive.ContentProps & {
    sameWidth?: boolean;
  };

  let className: $$Props["class"] = undefined;
  export let sideOffset: $$Props["sideOffset"] = 4;
  export let sameWidth: boolean = false;
  export { className as class };
</script>

<DropdownMenuPrimitive.Portal>
  <DropdownMenuPrimitive.Content
    {sideOffset}
    class={cn(
      "z-50 min-w-[8rem] rounded-md border bg-popover p-1.5 text-popover-foreground shadow-md focus:outline-none",
      sameWidth && "w-[var(--bits-floating-anchor-width)]",
      className,
    )}
    {...$$restProps}
  >
    <slot />
  </DropdownMenuPrimitive.Content>
</DropdownMenuPrimitive.Portal>
