<script lang="ts">
  import { cn } from "@rilldata/web-common/lib/shadcn";
  import { AlertDialog as AlertDialogPrimitive } from "bits-ui";

  type $$Props = AlertDialogPrimitive.TitleProps;

  let className: $$Props["class"] = undefined;
  export let level: $$Props["level"] = "h3";
  export { className as class };
</script>

<AlertDialogPrimitive.Title
  class={cn("text-lg text-fg-primary font-semibold", className)}
  {level}
  {...$$restProps}
>
  <slot />
</AlertDialogPrimitive.Title>
