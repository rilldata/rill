<script lang="ts">
  import { ContextMenu as ContextMenuPrimitive } from "bits-ui";
  import { cn } from "@rilldata/web-common/lib/shadcn";

  type $$Props = ContextMenuPrimitive.SubContentProps;

  let className: $$Props["class"] = undefined;
  export { className as class };
</script>

<ContextMenuPrimitive.SubContent
  class={cn(
    "z-50 min-w-[8rem] rounded-md border bg-popover p-1 text-popover-foreground shadow-lg focus:outline-none",
    className,
  )}
  {...$$restProps}
>
  <slot />
</ContextMenuPrimitive.SubContent>
