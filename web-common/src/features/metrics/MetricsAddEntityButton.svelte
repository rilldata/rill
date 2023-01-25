<script lang="ts">
  import AddIcon from "@rilldata/web-common/components/icons/Add.svelte";

  export let clickHandler;
  export let label: string;
</script>

<div
  class="w-full p-2 pl-4 pr-4 border-2 border-gray-200 bg-white"
  on:click={clickHandler}
>
  <div class="flex items-center">
    <div class="gap-x-2 pr-4">
      <AddIcon />
    </div>
    <div>{label}</div>
  </div>
</div>
