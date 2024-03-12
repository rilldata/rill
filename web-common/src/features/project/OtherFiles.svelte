<script lang="ts">
  import { page } from "$app/stores";
  import NavigationEntry from "../../layout/navigation/NavigationEntry.svelte";
  import NavigationHeader from "../../layout/navigation/NavigationHeader.svelte";

  let showOtherFiles = true;
</script>

<NavigationHeader bind:show={showOtherFiles} toggleText="others">
  Others
</NavigationHeader>

{#if showOtherFiles}
  <ol>
    <li aria-label={"rill.yaml"}>
      <NavigationEntry
        showContextMenu={false}
        name={"rill.yaml"}
        href={`/rill.yaml`}
        open={$page.url.pathname === `/rill.yaml`}
      />
    </li>
  </ol>
{/if}
