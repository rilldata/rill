<script lang="ts">
  import { cn } from "@rilldata/web-common/lib/shadcn";
  import { DropdownMenu as DropdownMenuPrimitive } from "bits-ui";
  import { Check, X } from "lucide-svelte";
  import type { Snippet } from "svelte";

  // svelte-ignore custom_element_props_identifier
  let {
    class: className,
    checked = $bindable(false),
    checkSize = "h-4 w-4",
    href,
    preloadData = true,
    showXForSelected = false,
    checkRight = false,
    closeOnSelect = true,
    children,
    ...restProps
  }: DropdownMenuPrimitive.CheckboxItemProps & {
    checkSize?: string;
    checkRight?: boolean;
    href?: string;
    preloadData?: boolean;
    showXForSelected?: boolean;
    closeOnSelect?: boolean;
    children?: Snippet;
  } = $props();
</script>

<svelte:element
  this={href ? "a" : "div"}
  {href}
  class="font-normal text-inherit"
  rel="noopener noreferrer"
  data-sveltekit-preload-data={preloadData ? "hover" : "false"}
>
  <DropdownMenuPrimitive.CheckboxItem
    bind:checked
    {closeOnSelect}
    class={cn(
      "relative flex cursor-pointer text-fg-primary select-none items-center rounded-sm py-1.5 px-2 gap-x-2 text-xs outline-none data-[highlighted]:bg-popover-accent data-[highlighted]:text-fg-accent data-[disabled]:pointer-events-none data-[disabled]:opacity-50 hover:bg-popover-accent hover:rounded-sm focus:bg-popover-accent focus:rounded-sm",
      className,
      checkRight && "flex-row-reverse justify-between",
    )}
    {...restProps}
  >
    <span class="flex flex-none h-3.5 w-3.5 items-center justify-center">
      {#if checked}
        {#if showXForSelected}
          <X class="{checkSize} text-fg-primary" />
        {:else}
          <Check class="{checkSize} text-fg-primary" />
        {/if}
      {/if}
    </span>
    {@render children?.()}
  </DropdownMenuPrimitive.CheckboxItem>
</svelte:element>
