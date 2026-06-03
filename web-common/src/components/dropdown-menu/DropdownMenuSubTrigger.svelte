<script lang="ts">
  import { cn } from "@rilldata/web-common/lib/shadcn";
  import { DropdownMenu as DropdownMenuPrimitive } from "bits-ui";
  import { ChevronRight } from "lucide-svelte";
  import type { Snippet } from "svelte";

  // svelte-ignore custom_element_props_identifier
  let {
    class: className,
    inset,
    children,
    ...restProps
  }: DropdownMenuPrimitive.SubTriggerProps & {
    inset?: boolean;
    children?: Snippet;
  } = $props();
</script>

<DropdownMenuPrimitive.SubTrigger
  class={cn(
    "flex cursor-pointer select-none items-center rounded-sm px-2 py-1.5 text-xs outline-none data-[highlighted]:bg-popover-accent data-[state=open]:bg-popover-accent data-[highlighted]:text-fg-accent data-[state=open]:text-fg-accent",
    inset && "pl-8",
    className,
  )}
  {...restProps}
>
  {@render children?.()}
  <ChevronRight class="ml-auto h-4 w-4" />
</DropdownMenuPrimitive.SubTrigger>
