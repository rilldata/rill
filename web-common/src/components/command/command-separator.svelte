<script lang="ts">
  import { cn } from "@rilldata/web-common/lib/shadcn";
  import { Command as CommandPrimitive } from "bits-ui";

  let className: string | undefined | null = undefined;
  export { className as class };
</script>

<CommandPrimitive.Separator
  class={cn("bg-border -mx-1 h-px", className)}
  {...$$restProps}
/>
