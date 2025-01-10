<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import Eye from "@rilldata/web-common/components/icons/Eye.svelte";
  import EyeInvisible from "@rilldata/web-common/components/icons/EyeInvisible.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { copyToClipboard } from "@rilldata/web-common/lib/actions/copy-to-clipboard";

  export let value: string;

  let showValue = false;
  let copied = false;

  $: isEmpty = value.length === 0;
  $: isValueHidden = !showValue;

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
</script>

<div class="flex flex-row gap-2 items-center truncate">
  <button
    class="hover:bg-slate-100 rounded-sm p-0.5 flex-none"
    on:click={toggleShowValue}
  >
    <svelte:component
      this={showValue ? EyeInvisible : Eye}
      color="#374151"
      size="18px"
    />
  </button>

  {#if showValue}
    <Tooltip distance={6} location="top" suppress={isValueHidden || isEmpty}>
      <button on:click={onCopy} class="truncate">
        <span
          class:italic={isEmpty}
          class="text-sm text-gray-800 font-medium truncate"
          class:cursor-pointer={showValue}
          title={value}
        >
          {value || "Empty"}
        </span>
      </button>
      <TooltipContent slot="tooltip-content">
        {copied ? "Copied!" : "Click to copy"}
      </TooltipContent>
    </Tooltip>
  {:else}
    <span class="pointer-events-none"> ••••••••••• </span>
  {/if}
</div>
