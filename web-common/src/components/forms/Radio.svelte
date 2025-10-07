<script lang="ts">
  export let value: string;
  export let options: Array<{
    value: string;
    label: string;
    description?: string;
    hint?: string;
  }> = [];
  export let disabled: boolean = false;
  export let name: string = "radio-group";

  function handleValueChange(newValue: string) {
    if (!disabled) {
      value = newValue;
    }
  }
</script>

<div class="flex flex-col gap-4">
  {#each options as option (option.value)}
    <label
      class="flex items-start gap-3 p-4 rounded-lg cursor-pointer transition-all hover:border-gray-300 hover:bg-gray-50 {disabled
        ? 'cursor-not-allowed opacity-50'
        : ''}"
      class:disabled
    >
      <input
        type="radio"
        {name}
        value={option.value}
        checked={value === option.value}
        {disabled}
        on:change={() => handleValueChange(option.value)}
        class="mt-1 w-4 h-4 text-blue-600 border-gray-300 focus:ring-blue-500"
      />
      <div class="flex-1">
        <div class="text-sm font-medium text-gray-900 mb-1">{option.label}</div>
        {#if option.description}
          <div class="text-sm text-gray-600 mb-2">{option.description}</div>
        {/if}
        {#if option.hint}
          <div class="text-xs text-gray-500">{option.hint}</div>
        {/if}
      </div>
    </label>
  {/each}
</div>
