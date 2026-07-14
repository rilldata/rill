<script lang="ts">
  import { Separator as SeparatorPrimitive } from "bits-ui";
  import { cn } from "@rilldata/web-common/lib/shadcn";

  type $$Props = SeparatorPrimitive.RootProps;

  let className: $$Props["class"] = undefined;
  export { className as class };
</script>

<SeparatorPrimitive.Root
  class={cn("-mx-1 my-1 h-px bg-surface-muted", className)}
  {...$$restProps}
/>
