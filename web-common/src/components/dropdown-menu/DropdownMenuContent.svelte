<script lang="ts">
  import { getThemeBoundaryClass } from "@rilldata/web-common/features/themes/theme-boundary";
  import { cn } from "@rilldata/web-common/lib/shadcn";
  import { DropdownMenu as DropdownMenuPrimitive } from "bits-ui";
  import type { Snippet } from "svelte";

  // svelte-ignore custom_element_props_identifier
  let {
    class: className,
    sideOffset = 4,
    sameWidth = false,
    children,
    ...restProps
  }: DropdownMenuPrimitive.ContentProps & {
    sameWidth?: boolean;
    children?: Snippet;
  } = $props();

  const themeBoundaryClass = getThemeBoundaryClass();
</script>

<DropdownMenuPrimitive.Portal>
  <DropdownMenuPrimitive.Content
    {sideOffset}
    class={cn(
      themeBoundaryClass,
      "z-50 min-w-[8rem] rounded-md border bg-popover p-1.5 text-popover-foreground shadow-md focus:outline-none",
      sameWidth && "w-[var(--bits-floating-anchor-width)]",
      className,
    )}
    {...restProps}
  >
    {@render children?.()}
  </DropdownMenuPrimitive.Content>
</DropdownMenuPrimitive.Portal>
