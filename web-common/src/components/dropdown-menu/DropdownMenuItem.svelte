<script lang="ts">
  import { cn } from "@rilldata/web-common/lib/shadcn";
  import { DropdownMenu as DropdownMenuPrimitive } from "bits-ui";
  import type { Snippet } from "svelte";

  type ItemType = "default" | "destructive";

  // svelte-ignore custom_element_props_identifier
  let {
    class: className,
    inset,
    type = "default" as ItemType,
    href,
    preloadData = true,
    children,
    ...restProps
  }: DropdownMenuPrimitive.ItemProps & {
    inset?: boolean;
    type?: ItemType;
    href?: string;
    target?: string;
    rel?: string;
    preloadData?: boolean;
    children?: Snippet;
  } = $props();
</script>

<svelte:element
  this={href ? "a" : "div"}
  {href}
  class="font-normal text-inherit"
  data-sveltekit-preload-data={preloadData ? "hover" : "false"}
>
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
</svelte:element>
