<script lang="ts">
  import EyeIcon from "@rilldata/web-common/components/icons/EyeIcon.svelte";
  import EyeOffIcon from "@rilldata/web-common/components/icons/EyeOffIcon.svelte";

  export let value: string;

  let showValue = false;

  $: inputType = showValue ? "text" : "password";

  $: isEmpty = value.length === 0;

  let displayValue = value;
  $: if (!showValue) {
    // 16 characters
    displayValue = "****************";
  } else {
    if (isEmpty) {
      displayValue = "Empty";
    } else {
      displayValue = value;
    }
  }

  function toggleShowValue() {
    showValue = !showValue;
  }
</script>

<div class="flex flex-row gap-[10px] items-center">
  <button on:click={toggleShowValue}>
    {#if !showValue}
      <EyeIcon color="#94A3B8" size="16" />
    {:else}
      <EyeOffIcon color="#94A3B8" size="16" />
    {/if}
  </button>
  <input
    readonly
    type={inputType}
    class="text-sm text-gray-800 font-medium {isEmpty ? 'italic' : ''}"
    value={displayValue}
  />
</div>
