<script>
  import { File } from "lucide-svelte";

  export let size = "1em";
  export let className = "";
</script>

<File {size} class="stroke-icon-muted {className}" />
