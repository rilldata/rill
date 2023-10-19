<script lang="ts">
  export let value: string;
  export let id: string;
  export let label: string;
</script>

<div class="flex flex-col gap-y-2">
  <label for={id} class="text-gray-800 text-sm font-medium">{label}</label>
  <!-- Get calendar day -->
  <input
    type="date"
    class="bg-white rounded-sm border border-gray-300 px-3 py-[5px] h-8 cursor-pointer focus:outline-blue-500 w-full text-xs"
    bind:value
  />
</div>
