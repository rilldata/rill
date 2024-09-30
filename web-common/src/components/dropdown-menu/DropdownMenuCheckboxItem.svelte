<script lang="ts">
  import { cn } from "@rilldata/web-common/lib/shadcn";
  import { DropdownMenu as DropdownMenuPrimitive } from "bits-ui";
  import { Check, X } from "lucide-svelte";

  type $$Props = DropdownMenuPrimitive.CheckboxItemProps & {
    checkSize?: string;
    // See: https://www.bits-ui.com/docs/components/dropdown-menu#dropdownmenucheckboxitem
    // Converts div to anchor tag
    href?: string;
    showXForSelected?: boolean;
  };
  // type $$Events = DropdownMenuPrimitive.CheckboxItemEvents;

  let className: $$Props["class"] = undefined;
  export let checked: $$Props["checked"] = undefined;
  export let checkSize: $$Props["checkSize"] = "h-4 w-4";
  export let href: $$Props["href"] = undefined;
  export let showXForSelected: $$Props["showXForSelected"] = false;
  export { className as class };

  const iconColor = "#15141A";
</script>

<svelte:element this={href ? "a" : "div"} {href} rel="noopener noreferrer">
  <DropdownMenuPrimitive.CheckboxItem
    bind:checked
    role="menuitem"
    class={cn(
      "relative flex cursor-default select-none items-center rounded-sm py-1.5 pl-8 pr-2 text-xs outline-none data-[highlighted]:bg-accent data-[highlighted]:text-accent-foreground data-[disabled]:pointer-events-none data-[disabled]:opacity-50 hover:bg-accent hover:rounded-sm focus:bg-accent focus:rounded-sm",
      className,
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
    <span
      class="absolute left-2.5 flex h-3.5 w-3.5 items-center justify-center"
    >
      <DropdownMenuPrimitive.CheckboxIndicator>
        <svelte:component
          this={showXForSelected ? X : Check}
          class={checkSize}
          color={iconColor}
        />
      </DropdownMenuPrimitive.CheckboxIndicator>
    </span>
    <slot />
  </DropdownMenuPrimitive.CheckboxItem>
</svelte:element>
