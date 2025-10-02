<script lang="ts">
  import { createEventDispatcher } from "svelte";

  export let value: string | boolean;
  export let options: Array<{ value: string | boolean; label: string }> = [];
  export let disabled: boolean = false;

  const dispatch = createEventDispatcher();

  function handleChange(event: Event) {
    const target = event.target as HTMLSelectElement;
    value = target.value;
    dispatch("change", { value: target.value });
  }
</script>

<div class="text-sm font-medium mb-2">Connector type</div>
<select
  id="managed"
  bind:value
  on:change={handleChange}
  class="form-select"
  {disabled}
>
  {#each options as option}
    <option value={option.value}>{option.label}</option>
  {/each}
</select>
