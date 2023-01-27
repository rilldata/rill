<script lang="ts">
  export let connector: string;

  $: connector = connector == "local_file" ? "local" : connector;
</script>

<div
  class="w-max text-gray-600 bg-gray-200  items-center px-1 rounded flex gap-x-1 "
  style:font-size="10px"
>
  <span>{connector}</span>
</div>
