<script lang="ts">
  import { Tabs as TabsPrimitive } from "bits-ui";
  import { cn } from "../../lib/shadcn";

  type $$Props = TabsPrimitive.ListProps;

  let className: $$Props["class"] = undefined;
  export { className as class };
</script>

<TabsPrimitive.List
  class={cn("inline-flex h-9 items-center justify-start gap-x-4", className)}
  {...$$restProps}
>
  <slot />
</TabsPrimitive.List>
