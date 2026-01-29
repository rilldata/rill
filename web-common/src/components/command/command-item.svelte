<script lang="ts">
  import { cn } from "@rilldata/web-common/lib/shadcn";
  import { Command as CommandPrimitive } from "cmdk-sv";

  // type $$Props = CommandPrimitive.ItemProps;

  export let asChild = false;

  let className: string | undefined | null = undefined;
  export { className as class };
</script>

<CommandPrimitive.Item
  {asChild}
  class={cn(
    "aria-selected:bg-popover-accent aria-selected:text-fg-accent relative flex cursor-default select-none items-center rounded-sm px-2 py-1.5 text-xs outline-none data-[disabled]:pointer-events-none data-[disabled]:opacity-50",
    className,
  )}
  {...$$restProps}
  let:action
  let:attrs
>
  <slot {action} {attrs} />
</CommandPrimitive.Item>
