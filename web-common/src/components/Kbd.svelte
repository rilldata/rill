<script lang="ts">
  import { cn } from "@rilldata/web-common/lib/shadcn";

  export let className = "";
</script>

<kbd
  class={cn(
    "px-1.5 py-1 bg-surface-subtle rounded font-mono border border-gray-300 dark:border-gray-500",
    className,
  )}
>
  <slot />
</kbd>
