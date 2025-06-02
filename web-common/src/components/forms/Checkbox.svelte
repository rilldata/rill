<script lang="ts">
  export let checked: boolean;
  export let id: string;
  export let label: string;
  export let inverse = false;
  export let disabled = false;
  export let onCheckedChange: (checked: boolean) => void = () => {};
</script>

<div class="flex gap-x-2 {inverse ? 'flex-row-reverse justify-end' : ''}">
  <label for={id} class="text-gray-600">{label}</label>
  <input
    type="checkbox"
    {id}
    {disabled}
    bind:checked
    on:change={() => onCheckedChange(checked)}
  />
</div>
