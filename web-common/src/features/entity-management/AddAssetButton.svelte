<script lang="ts">
  import Add from "../../components/icons/Add.svelte";

  export let id: string;
  export let label: string;
  export let bold: boolean;
</script>

<button
  on:click
  {id}
  class="pl-2.5 pr-2 py-1 hover:bg-gray-100 w-full text-left flex gap-x-1.5 items-center text-xs {bold &&
    'font-bold'}"
>
  <Add className="text-gray-900 bg-gray-300 rounded" size={"16px"} />
  {label}
</button>
