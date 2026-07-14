<script lang="ts">
  import type { Snippet } from "svelte";
  let { onclick, children }: { onclick?: () => void; children?: Snippet } =
    $props();
</script>

<button type="button" {onclick}>
  {@render children?.()}
</button>
