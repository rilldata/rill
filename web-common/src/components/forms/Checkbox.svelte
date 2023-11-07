<script lang="ts">
  export let checked: boolean;
  export let id: string;
  export let label: string;
</script>

<div class="flex gap-x-2">
  <label for={id} class="text-gray-600">{label}</label>
  <input type="checkbox" {id} bind:checked />
</div>
