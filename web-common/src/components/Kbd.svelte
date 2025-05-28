<script lang="ts">
  import { cn } from "@rilldata/web-common/lib/shadcn";

  export let className = "";
</script>

<kbd
  class={cn(
    "px-1 py-1 bg-gray-200 dark:bg-gray-600 rounded font-mono border border-gray-300 dark:border-gray-500",
    className,
  )}
>
  <slot />
</kbd>
