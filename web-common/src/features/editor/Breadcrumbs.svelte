<script lang="ts">
  import CaretDownIcon from "../../components/icons/CaretDownIcon.svelte";

  export let filePath: string;

  $: breadcrumbs = filePath.split("/").map((_, i, arr) => {
    const path = arr.slice(0, i + 1).join("/");
    return {
      path,
      name: path.split("/").pop(),
    };
  });
</script>

{#if breadcrumbs}
  <h1
    style:font-size="16px"
    class="w-full overflow-x-hidden grid grid-flow-col justify-start items-center gap-x-1"
  >
    {#each breadcrumbs as { path, name }, i}
      {name}
      {#if i < breadcrumbs.length - 1}
        <CaretDownIcon size="16px" className="transform -rotate-90" />
      {/if}
    {/each}
  </h1>
{/if}
