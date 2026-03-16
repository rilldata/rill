<script lang="ts">
  import { cn } from "@rilldata/web-common/lib/shadcn";
  import { DropdownMenu as DropdownMenuPrimitive } from "bits-ui";
  import type { Snippet } from "svelte";

  type ItemType = "default" | "destructive";

  let {
    class: className,
    inset,
    type = "default" as ItemType,
    children,
    ...restProps
  }: DropdownMenuPrimitive.ItemProps & {
    inset?: boolean;
    type?: ItemType;
    href?: string;
    children?: Snippet;
  } = $props();
</script>

<DropdownMenuPrimitive.Item
  class={cn(
    "relative flex gap-x-2 text-fg-primary cursor-default select-none items-center rounded-sm px-2 py-1.5 text-xs outline-none data-[highlighted]:bg-surface-hover data-[highlighted]:text-fg-accent data-[highlighted]:cursor-pointer data-[disabled]:pointer-events-none data-[disabled]:opacity-50",
    inset && "pl-8",
    type === "destructive" &&
      "text-red-500 hover:text-red-600 data-[highlighted]:text-red-600",
    className,
  )}
  {...restProps}
>
  {@render children?.()}
</DropdownMenuPrimitive.Item>
