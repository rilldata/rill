<script lang="ts">
  import { cn } from "@rilldata/web-common/lib/shadcn";

  // type $$Props = HTMLAttributes<HTMLSpanElement>;

  let className: string | undefined | null = undefined;
  export { className as class };
</script>

<span
  class={cn("text-fg-secondary ml-auto text-xs tracking-widest", className)}
  {...$$restProps}
>
  <slot />
</span>
