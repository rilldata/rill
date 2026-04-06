<script lang="ts">
  import { cn } from "@rilldata/web-common/lib/shadcn";
  import { DropdownMenu as DropdownMenuPrimitive } from "bits-ui";
  import type { Snippet } from "svelte";

  // svelte-ignore custom_element_props_identifier
  let {
    class: className,
    sideOffset = 2,
    align = "start" as const,
    children,
    ...restProps
  }: DropdownMenuPrimitive.SubContentProps & {
    children?: Snippet;
  } = $props();
</script>

<DropdownMenuPrimitive.SubContent
  {sideOffset}
  {align}
  class={cn(
    "z-50 min-w-[8rem] rounded-md border bg-popover p-1.5 text-popover-foreground shadow-md focus:outline-none",
    className,
  )}
  {...restProps}
>
  {@render children?.()}
</DropdownMenuPrimitive.SubContent>
