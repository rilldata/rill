<script lang="ts">
  import { cn } from "@rilldata/web-common/lib/shadcn";

  import type { HTMLAttributes } from "svelte/elements";

  type $$Props = HTMLAttributes<HTMLSpanElement>;

  let className: $$Props["class"] = undefined;
  export { className as class };
</script>

<span
  class={cn("ml-auto text-xs tracking-widest text-fg-secondary", className)}
  {...$$restProps}
>
  <slot />
</span>
