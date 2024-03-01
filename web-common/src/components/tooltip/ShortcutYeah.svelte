<script lang="ts">
  import ShiftKey from "./ShiftKey.svelte";
  import MetaKey from "./MetaKey.svelte";

  export let modifiers: Array<"ctrl" | "shift" | "option" | "meta"> = [];
  export let keys: string[] = [];
  export let click = true;
</script>

<span class=" text-gray-400">
  {#each modifiers as modifier}
    {#if modifier === "ctrl"}
      Ctrl
    {/if}
    {#if modifier === "shift"}
      <ShiftKey />
    {/if}
    {#if modifier === "option"}
      Option
    {/if}
    {#if modifier === "meta"}
      <MetaKey />
    {/if}

    +
  {/each}
  {#if keys.length > 0}
    {#each keys as key, i}
      {key}
      {i < keys.length - 1 ? " + " : ""}
    {/each}
    +
  {/if}
  {#if click}
    Click
  {/if}
</span>
