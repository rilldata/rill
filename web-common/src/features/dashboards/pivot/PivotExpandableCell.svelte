<script lang="ts">
  import CheckBox from "@rilldata/web-common/components/icons/CheckBox.svelte";

  export let row;
  export let value;
</script>

<div class="flex gap-x-1" style:padding-left={`${row.depth * 2}rem`}>
  {#if row.getCanExpand()}
    <button on:click={row.getToggleExpandedHandler()} class="cursor-pointer">
      {row.getIsExpanded() ? "▼" : "▶"}
    </button>
  {:else}
    🔵
  {/if}
  {value}
</div>
