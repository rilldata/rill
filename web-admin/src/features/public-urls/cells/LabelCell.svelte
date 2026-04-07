<script lang="ts">
  import { ExternalLinkIcon } from "lucide-svelte";

  export let displayName: string;
  export let dashboardTitle: string;
  export let url: string;
</script>

<a
  href={url}
  target="_blank"
  rel="noopener noreferrer"
  class="flex items-center gap-x-2 hover:text-accent-primary-action truncate"
>
  <ExternalLinkIcon size={14} class="shrink-0" />
  <span class="truncate">
    {displayName || dashboardTitle || "Untitled"}
  </span>
</a>
