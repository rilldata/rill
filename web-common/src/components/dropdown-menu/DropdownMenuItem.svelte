<script lang="ts">
  import { cn } from "@rilldata/web-common/lib/shadcn";
  import { DropdownMenu as DropdownMenuPrimitive } from "bits-ui";

  type ItemType = "default" | "destructive";

  type $$Props = DropdownMenuPrimitive.ItemProps & {
    inset?: boolean;
    type?: ItemType; // Add type prop
  };
  // type $$Events = DropdownMenuPrimitive.ItemEvents;

  let className: $$Props["class"] = undefined;
  export let inset: $$Props["inset"] = undefined;
  export let type: $$Props["type"] = "default"; // Export type prop with default value
  export { className as class };
</script>

<DropdownMenuPrimitive.Item
  class={cn(
    "relative flex cursor-default select-none items-center rounded-sm px-2 py-1.5 text-xs outline-none data-[highlighted]:bg-accent data-[highlighted]:text-accent-foreground data-[highlighted]:cursor-pointer data-[disabled]:pointer-events-none data-[disabled]:opacity-50",
    inset && "pl-8",
    type === "destructive" &&
      "text-red-500 hover:text-red-600 data-[highlighted]:text-red-600",
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
  <slot />
</DropdownMenuPrimitive.Item>
