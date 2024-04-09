<script lang="ts">
  import { slide } from "svelte/transition";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import Docs from "@rilldata/web-common/components/icons/Docs.svelte";

  type DirectoryOrFile = Map<string, DirectoryOrFile | null>;

  export let name: string | undefined = "root";
  export let expanded = false;
  export let files: DirectoryOrFile;
  export let parent: string = "/";
  export let depth = 0;

  function toggle() {
    expanded = !expanded;
  }
  const ICON_SIZE = 14;
  const BASE_DEPTH = 4;
</script>

<button
  class:expanded
  on:click={toggle}
  style:padding-left="{depth * ICON_SIZE + BASE_DEPTH}px"
>
  <span class:rotate-0={expanded} class="transition-transform -rotate-90">
    <CaretDownIcon size="{ICON_SIZE}px" className="fill-gray-400" />
  </span>
  {name}
</button>

{#if expanded}
  <ul transition:slide={{ duration: 300 }}>
    {#each files as [name, directory]}
      {@const newDepth = depth + 1}
      <li>
        {#if directory}
          <svelte:self
            {name}
            files={directory}
            parent={parent + name + "/"}
            depth={newDepth}
          />
        {:else}
          <a
            href={`/file${parent + name}`}
            style:padding-left="{newDepth * ICON_SIZE + BASE_DEPTH}px"
          >
            <Docs className="fill-gray-400" size="{ICON_SIZE}px" />
            {name}
          </a>
        {/if}
      </li>
    {/each}
  </ul>
{/if}

<style lang="postcss">
  a,
  button {
    @apply h-6;
  }

  button,
  a {
    @apply flex gap-x-1 items-center font-medium w-full text-black;
  }

  a:hover,
  button:hover {
    @apply bg-gray-100;
  }
</style>
