<script lang="ts">
  import { page } from "$app/stores";
  import { slide } from "svelte/transition";
  import { LIST_SLIDE_DURATION } from "../../layout/config";
  import NavigationEntry from "../../layout/navigation/NavigationEntry.svelte";
  import NavigationHeader from "../../layout/navigation/NavigationHeader.svelte";

  let showOtherFiles = true;
</script>

<NavigationHeader bind:show={showOtherFiles} toggleText="others">
  Others
</NavigationHeader>

{#if showOtherFiles}
  <div class="pb-3" transition:slide={{ duration: LIST_SLIDE_DURATION }}>
    <!-- rill.yaml -->
    <NavigationEntry
      showContextMenu={false}
      name={"rill.yaml"}
      href={`/rill.yaml`}
      open={$page.url.pathname === `/rill.yaml`}
    />
  </div>
{/if}
