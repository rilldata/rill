<script lang="ts">
  export let value: string;
  export let id: string;
  export let label: string;
</script>

<div>
  <label for={id} class="text-gray-600">{label}</label>
  <div class="flex gap-x-2">
    <!-- Get calendar day -->
    <input
      type="date"
      class="bg-white rounded-sm border border-gray-300 px-3 py-[5px] h-8 cursor-pointer focus:outline-blue-500 w-full text-xs"
      bind:value
    />
    <!-- Get time -->
    <input
      type="time"
      class="bg-white rounded-sm border border-gray-300 px-3 py-[5px] h-8 cursor-pointer focus:outline-blue-500 w-full text-xs"
      bind:value
    />
  </div>
</div>
