<script lang="ts">
  import { ZoomOut, ZoomIn } from "lucide-svelte";
</script>

<button>
  <ZoomOut size="16px" />
</button>
<button>
  <ZoomIn size="16px" />
</button>

<style lang="postcss">
  button {
    @apply w-8 flex items-center justify-center;
  }
</style>
