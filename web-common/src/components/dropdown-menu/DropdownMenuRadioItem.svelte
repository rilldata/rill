<script lang="ts">
  import { cn } from "@rilldata/web-common/lib/shadcn";
  import { DropdownMenu as DropdownMenuPrimitive } from "bits-ui";
  import { Circle } from "lucide-svelte";

  type $$Props = DropdownMenuPrimitive.RadioItemProps;

  let className: $$Props["class"] = undefined;
  export let value: $$Props["value"];
  export { className as class };
</script>

<DropdownMenuPrimitive.RadioItem
  class={cn(
    "group relative flex cursor-default select-none items-center rounded-sm py-1.5 pl-8 pr-2 text-sm outline-none data-[highlighted]:bg-popover-accent data-[highlighted]:text-fg-accent data-[disabled]:pointer-events-none data-[disabled]:opacity-50",
    className,
  )}
  {value}
  {...$$restProps}
>
  <span class="absolute left-2 flex h-3.5 w-3.5 items-center justify-center">
    <Circle
      class="h-2 w-2 fill-current hidden group-data-[state=checked]:block"
    />
  </span>
  <slot />
</DropdownMenuPrimitive.RadioItem>
