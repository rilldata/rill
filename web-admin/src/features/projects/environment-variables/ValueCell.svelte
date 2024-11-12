<script lang="ts">
  import EyeIcon from "@rilldata/web-common/components/icons/EyeIcon.svelte";
  import EyeOffIcon from "@rilldata/web-common/components/icons/EyeOffIcon.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";

  export let value: string;

  let showValue = false;
  let copied = false;
  let displayValue = value;

  $: inputType = showValue ? "text" : "password";
  $: isEmpty = value.length === 0;
  $: isValueHidden = !showValue;

  // 16 characters
  const REDACTED_VALUE = "****************";

  $: if (!showValue) {
    displayValue = REDACTED_VALUE;
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

  function copyToClipboard(text: string) {
    navigator.clipboard.writeText(text).catch(console.error);
  }

  function onCopy() {
    if (isValueHidden) {
      return;
    }

    copyToClipboard(value);
    copied = true;

    setTimeout(() => {
      copied = false;
    }, 2_000);
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
  <Tooltip distance={8} location="top" suppress={isValueHidden || isEmpty}>
    <input
      readonly
      type={inputType}
      class="text-sm text-gray-800 font-medium {isEmpty
        ? 'italic'
        : ''} outline-none"
      class:cursor-pointer={showValue}
      value={displayValue}
      title={showValue ? displayValue : ""}
      on:click={onCopy}
    />
    <TooltipContent maxWidth="600px" slot="tooltip-content">
      {copied ? "Copied!" : "Click to copy"}
    </TooltipContent>
  </Tooltip>
</div>
