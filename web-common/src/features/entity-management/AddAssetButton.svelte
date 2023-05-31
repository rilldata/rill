<script lang="ts">
  import Add from "../../components/icons/Add.svelte";

  export let id: string;
  export let label: string;
</script>

<button
  on:click
  {id}
  class="py-1 pl-2.5 pr-2 hover:bg-gray-100 w-full text-left flex gap-x-2 items-center text-xs"
>
  <Add className="text-gray-900 bg-gray-300 rounded" size={"16px"} />
  {label}
</button>
