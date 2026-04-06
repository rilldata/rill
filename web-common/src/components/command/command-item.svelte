<script lang="ts">
  import { cn } from "@rilldata/web-common/lib/shadcn";
  import { Command as CommandPrimitive } from "bits-ui";

  let className: string | undefined | null = undefined;
  export { className as class };
</script>

<CommandPrimitive.Item
  class={cn(
    "hover:bg-popover-accent hover:text-fg-accent data-[selected=true]:bg-popover-accent data-[selected=true]:text-fg-accent relative flex cursor-default select-none items-center rounded-sm px-2 py-1.5 text-xs outline-none data-[disabled]:pointer-events-none data-[disabled]:opacity-50",
    className,
  )}
  {...$$restProps}
>
  <slot />
</CommandPrimitive.Item>
