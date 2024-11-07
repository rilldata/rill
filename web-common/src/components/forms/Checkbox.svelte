<script lang="ts">
  export let checked: boolean;
  export let id: string;
  export let label: string;
  export let inverse = false;
  export let onCheckedChange: (checked: boolean) => void = () => {};
</script>

<div class="flex gap-x-2 {inverse ? 'flex-row-reverse' : ''}">
  <label for={id} class="text-gray-600">{label}</label>
  <input
    type="checkbox"
    {id}
    bind:checked
    on:change={() => onCheckedChange(checked)}
  />
</div>
