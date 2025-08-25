<script lang="ts">
  import { cn } from "@rilldata/web-common/lib/shadcn";
  import { DropdownMenu as DropdownMenuPrimitive } from "bits-ui";
  import { Check, X } from "lucide-svelte";

  type $$Props = DropdownMenuPrimitive.CheckboxItemProps & {
    checkSize?: string;
    checkRight?: boolean;
    // See: https://www.bits-ui.com/docs/components/dropdown-menu#dropdownmenucheckboxitem
    // Converts div to anchor tag
    href?: string;
    preloadData?: boolean;
    showXForSelected?: boolean;
  };
  // type $$Events = DropdownMenuPrimitive.CheckboxItemEvents;

  let className: $$Props["class"] = undefined;
  export let checked: $$Props["checked"] = undefined;
  export let checkSize: $$Props["checkSize"] = "h-4 w-4";
  export let href: $$Props["href"] = undefined;
  export let preloadData: $$Props["preloadData"] = true;
  export let showXForSelected: $$Props["showXForSelected"] = false;
  export let checkRight: $$Props["checkRight"] = false;
  export { className as class };

  const iconColor = "var(--color-gray-800)";
</script>

<svelte:element
  this={href ? "a" : "div"}
  {href}
  rel="noopener noreferrer"
  data-sveltekit-preload-data={preloadData ? "hover" : "false"}
>
  <DropdownMenuPrimitive.CheckboxItem
    {checked}
    role="menuitem"
    class={cn(
      "relative flex cursor-pointer select-none items-center rounded-sm py-1.5 px-2 gap-x-2 text-xs outline-none data-[highlighted]:bg-accent data-[highlighted]:text-accent-foreground data-[disabled]:pointer-events-none data-[disabled]:opacity-50 hover:bg-accent hover:rounded-sm focus:bg-accent focus:rounded-sm",
      className,
      checkRight && "flex-row-reverse justify-between",
    )}
    {...$$restProps}
    on:click
    on:keydown
    on:focusin
    on:focusout
    on:pointerdown
    on:pointerleave
    on:pointermove
  >
    <span class="flex flex-none h-6 w-6 items-center justify-center rounded-sm hover:bg-gray-100 transition-colors">
      {#if checked}
        <svelte:component
          this={showXForSelected ? X : Check}
          class={checkSize}
          color={iconColor}
        />
      {:else}
        <!-- Invisible placeholder to maintain consistent spacing and clickable area -->
        <span class="h-4 w-4 opacity-0" aria-hidden="true"></span>
      {/if}
    </span>
    <slot />
  </DropdownMenuPrimitive.CheckboxItem>
</svelte:element>
