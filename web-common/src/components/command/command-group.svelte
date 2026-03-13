<script lang="ts">
  import { cn } from "@rilldata/web-common/lib/shadcn";
  import { Command as CommandPrimitive } from "bits-ui";

  let className: string | undefined | null = undefined;
  export { className as class };
</script>

<CommandPrimitive.Group
  class={cn(
    "text-fg-primary overflow-hidden py-1",
    className,
  )}
  {...$$restProps}
>
  <slot />
</CommandPrimitive.Group>
