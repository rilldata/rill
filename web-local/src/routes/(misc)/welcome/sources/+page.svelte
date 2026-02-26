<script lang="ts">
  import { goto } from "$app/navigation";
  import NewSourceSelector from "@rilldata/web-common/features/welcome/new-sources/NewSourceSelector.svelte";
</script>

<div class="min-w-[900px]">
  <NewSourceSelector
    startConnectorSelection={(name) => void goto(`/welcome/sources/${name}`)}
  />
</div>
