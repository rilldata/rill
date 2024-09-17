<script lang="ts">
  import { cn } from "@rilldata/web-common/lib/shadcn";
  import type { HTMLTableAttributes } from "svelte/elements";

  type $$Props = HTMLTableAttributes;

  let className: $$Props["class"] = undefined;
  export { className as class };

  export let wrapperClass: $$Props["class"] = undefined;
</script>

<div class={cn("relative w-full overflow-auto", wrapperClass)}>
  <table
    class={cn("w-full caption-bottom text-sm", className)}
    {...$$restProps}
  >
    <slot />
  </table>
</div>
