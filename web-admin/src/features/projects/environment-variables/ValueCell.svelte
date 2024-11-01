<script lang="ts">
  import EyeIcon from "@rilldata/web-common/components/icons/EyeIcon.svelte";
  import EyeInvisible from "@rilldata/web-common/components/icons/EyeInvisible.svelte";
  import EyeOffIcon from "@rilldata/web-common/components/icons/EyeOffIcon.svelte";

  export let value: string;

  let showValue = false;

  $: inputType = showValue ? "text" : "password";

  let displayValue = value;
  $: if (!showValue) {
    // 16 characters
    displayValue = "****************";
  } else {
    displayValue = value;
  }
</script>

<div class="flex flex-row gap-[10px] items-center">
  <button on:click={() => (showValue = !showValue)}>
    {#if !showValue}
      <EyeIcon color="#94A3B8" size="16" />
    {:else}
      <EyeOffIcon color="#94A3B8" size="16" />
    {/if}
  </button>
  <input
    readonly
    type={inputType}
    class="text-sm text-gray-800 font-medium"
    value={displayValue}
  />
</div>
