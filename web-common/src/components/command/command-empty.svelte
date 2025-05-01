<script lang="ts">
  import { cn } from "@rilldata/web-common/lib/shadcn";
  import { Command as CommandPrimitive } from "cmdk-sv";

  // type $$Props = CommandPrimitive.EmptyProps;

  let className: string | undefined | null = undefined;
  export { className as class };
</script>

<CommandPrimitive.Empty
  class={cn("py-6 text-center text-xs", className)}
  {...$$restProps}
>
  <slot />
</CommandPrimitive.Empty>
