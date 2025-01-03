<script lang="ts">
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import Eye from "@rilldata/web-common/components/icons/Eye.svelte";
  import EyeInvisible from "@rilldata/web-common/components/icons/EyeInvisible.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { copyToClipboard } from "@rilldata/web-common/lib/actions/copy-to-clipboard";

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

  function onCopy() {
    if (isValueHidden) {
      return;
    }

    copyToClipboard(value, undefined, false);
    copied = true;

    setTimeout(() => {
      copied = false;
    }, 2_000);
  }

  $: title = displayValue !== "Empty" ? (showValue ? displayValue : "") : "";
</script>

<div class="flex flex-row gap-[10px] items-center">
  <button on:click={toggleShowValue}>
    {#if !showValue}
      <IconButton disableHover>
        <Eye color="#374151" size="16px" />
      </IconButton>
    {:else}
      <IconButton disableHover>
        <EyeInvisible color="#374151" size="16px" />
      </IconButton>
    {/if}
  </button>
  <Tooltip distance={6} location="top" suppress={isValueHidden || isEmpty}>
    <div class="w-fit">
      {#if inputType === "password"}
        <input
          readonly
          type="password"
          class="text-sm text-gray-800 font-medium {isEmpty
            ? 'italic'
            : ''} outline-none"
          class:cursor-pointer={showValue}
          value={displayValue}
          {title}
          on:click={onCopy}
        />
      {:else}
        <button on:click={onCopy} class="truncate max-w-[160.5px]">
          <span
            class="text-sm text-gray-800 font-medium {isEmpty
              ? 'italic'
              : ''} outline-none"
            class:cursor-pointer={showValue}
            {title}
          >
            {displayValue}
          </span>
        </button>
      {/if}
    </div>
    <TooltipContent slot="tooltip-content">
      {copied ? "Copied!" : "Click to copy"}
    </TooltipContent>
  </Tooltip>
</div>
