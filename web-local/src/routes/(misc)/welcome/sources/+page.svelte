<script lang="ts">
  import { goto } from "$app/navigation";
  import NewSourceSelector from "@rilldata/web-common/features/add-data/form/NewSourceSelector.svelte";
</script>

<div class="min-w-[900px]">
  <NewSourceSelector
    startConnectorSelection={(name) => void goto(`/welcome/sources/${name}`)}
  />
</div>
