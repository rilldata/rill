<script lang="ts">
  import { cn } from "@rilldata/web-common/lib/shadcn";
  import { DropdownMenu as DropdownMenuPrimitive } from "bits-ui";

  type $$Props = DropdownMenuPrimitive.SubContentProps;

  let className: $$Props["class"] = undefined;
  export { className as class };
</script>

<DropdownMenuPrimitive.SubContent
  class={cn(
    "z-50 min-w-[8rem] rounded-md border bg-popover p-1.5 text-popover-foreground shadow-md focus:outline-none",
    className,
  )}
  {...$$restProps}
>
  <slot />
</DropdownMenuPrimitive.SubContent>
