<script lang="ts">
  import { Select as SelectPrimitive } from "bits-ui";
  import { cn } from "@rilldata/web-common/lib/shadcn";

  type $$Props = SelectPrimitive.GroupHeadingProps;

  let className: $$Props["class"] = undefined;
  export { className as class };
</script>

<SelectPrimitive.GroupHeading
  class={cn("px-2 py-1.5 text-sm font-semibold", className)}
  {...$$restProps}
>
  <slot />
</SelectPrimitive.GroupHeading>
