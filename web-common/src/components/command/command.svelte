<script lang="ts">
  import { Command as CommandPrimitive } from "bits-ui";
  import { cn } from "@rilldata/web-common/lib/shadcn";

  type $$Props = {
    value?: string;
    class?: string | null;
  } & Record<string, unknown>;

  export let value: $$Props["value"] = undefined;

  let className: string | undefined | null = undefined;
  export { className as class };
</script>

<CommandPrimitive.Root
  class={cn(
    "bg-popover text-popover-foreground flex h-full w-full flex-col overflow-hidden rounded-md",
    className,
  )}
  bind:value
  {...$$restProps}
>
  <slot />
</CommandPrimitive.Root>
