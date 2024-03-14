<script lang="ts">
  import { Label as LabelPrimitive } from "bits-ui";
  import { cn } from "@rilldata/web-common/lib/shadcn";

  type $$Props = LabelPrimitive.Props;

  let className: $$Props["class"] = undefined;
  export { className as class };
</script>

<LabelPrimitive.Root
  {...$$restProps}
  class={cn(
    "text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70",
    className,
  )}
>
  <slot />
</LabelPrimitive.Root>
