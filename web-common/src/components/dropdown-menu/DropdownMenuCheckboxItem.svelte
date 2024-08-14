<script lang="ts">
  import { cn } from "@rilldata/web-common/lib/shadcn";
  import { DropdownMenu as DropdownMenuPrimitive } from "bits-ui";
  import { Check } from "lucide-svelte";

  type $$Props = DropdownMenuPrimitive.CheckboxItemProps & {
    checkSize?: string;
  };
  // type $$Events = DropdownMenuPrimitive.CheckboxItemEvents;

  let className: $$Props["class"] = undefined;
  export let checked: $$Props["checked"] = undefined;
  export let checkSize: $$Props["checkSize"] = "h-4 w-4";
  export { className as class };
</script>

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
  <span class="absolute left-2 flex h-3.5 w-3.5 items-center justify-center">
    <DropdownMenuPrimitive.CheckboxIndicator>
      <Check class={checkSize} color="#15141A" />
    </DropdownMenuPrimitive.CheckboxIndicator>
  </span>
  <slot />
</DropdownMenuPrimitive.CheckboxItem>
