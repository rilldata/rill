<script lang="ts">
  export let value: string | boolean;
  export let options: Array<{ value: string | boolean; label: string }> = [];
  export let disabled: boolean = false;
</script>

<div class="text-sm font-medium mb-2">Connector type</div>
<select id="managed" bind:value class="form-select" {disabled}>
  {#each options as option}
    <option value={option.value}>{option.label}</option>
  {/each}
</select>
