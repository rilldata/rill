<script lang="ts">
  import { createEventDispatcher } from "svelte";

  export let content;
  const dispatch = createEventDispatcher();
</script>

<!-- completely empty case -->
<div>Build a config off of a model or start with a skeleton.</div>
