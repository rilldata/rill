<script lang="ts">
  import { ContextMenu as ContextMenuPrimitive } from "bits-ui";
  import { cn } from "@rilldata/web-common/lib/shadcn";
  import { Dot } from "lucide-svelte";

  type $$Props = ContextMenuPrimitive.RadioItemProps;

  let className: $$Props["class"] = undefined;
  export let value: $$Props["value"];
  export { className as class };
</script>

<ContextMenuPrimitive.RadioItem
  class={cn(
    "group relative flex cursor-default select-none items-center rounded-sm py-1.5 pl-8 pr-2 text-sm outline-none data-[disabled]:pointer-events-none data-[highlighted]:bg-popover-accent data-[highlighted]:text-fg-accent data-[disabled]:opacity-50",
    className,
  )}
  {value}
  {...$$restProps}
>
  <span class="absolute left-2 flex h-3.5 w-3.5 items-center justify-center">
    <Dot class="h-4 w-4 fill-current hidden group-data-[state=checked]:block" />
  </span>
  <slot />
</ContextMenuPrimitive.RadioItem>
