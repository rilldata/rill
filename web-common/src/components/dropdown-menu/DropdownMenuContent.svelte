<script lang="ts">
  import { cn } from "@rilldata/web-common/lib/shadcn";
  import { DropdownMenu as DropdownMenuPrimitive } from "bits-ui";
  import type { Snippet } from "svelte";

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
</script>

<DropdownMenuPrimitive.Portal>
  <DropdownMenuPrimitive.Content
    {sideOffset}
    class={cn(
      "z-50 min-w-[8rem] rounded-md border bg-popover p-1.5 text-popover-foreground shadow-md focus:outline-none",
      sameWidth && "w-[var(--bits-floating-anchor-width)]",
      className,
    )}
    {...restProps}
  >
    {@render children?.()}
  </DropdownMenuPrimitive.Content>
</DropdownMenuPrimitive.Portal>
