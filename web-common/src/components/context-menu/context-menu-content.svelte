<script lang="ts">
  import { ContextMenu as ContextMenuPrimitive } from "bits-ui";

  import { getThemeBoundaryClass } from "@rilldata/web-common/features/themes/theme-boundary";
  import { cn } from "@rilldata/web-common/lib/shadcn";

  type $$Props = ContextMenuPrimitive.ContentProps;

  let className: $$Props["class"] = undefined;
  export { className as class };

  const themeBoundaryClass = getThemeBoundaryClass();
</script>

<ContextMenuPrimitive.Portal>
  <ContextMenuPrimitive.Content
    class={cn(
      themeBoundaryClass,
      "z-50 min-w-[8rem] rounded-md border bg-popover p-1 text-popover-foreground shadow-md focus:outline-none",
      className,
    )}
    {...$$restProps}
  >
    <slot />
  </ContextMenuPrimitive.Content>
</ContextMenuPrimitive.Portal>
