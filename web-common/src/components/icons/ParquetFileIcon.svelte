<script>
  import { Sheet } from "lucide-svelte";

  export let size = "1em";
  export let className = "";
</script>

<Sheet {size} class="stroke-icon-muted {className}" />
