<script lang="ts">
  import { copyToClipboard } from "@rilldata/web-common/lib/actions/copy-to-clipboard";
  import { EnvironmentType } from "./types";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";

  export let environment: string;
  export let name: string;

  // NOTE: if environment is empty, the variable is shared for all environments
  function getEnvironmentType(environment: string) {
    if (environment === EnvironmentType.UNDEFINED) {
      return "Development, Production";
    }
    if (environment === EnvironmentType.DEVELOPMENT) {
      return "Development";
    }
    if (environment === EnvironmentType.PRODUCTION) {
      return "Production";
    }
    return "";
  }

  $: environmentLabel = getEnvironmentType(environment);

  let copied = false;
  function onCopy() {
    copyToClipboard(name, undefined, false);
    copied = true;

    setTimeout(() => {
      copied = false;
    }, 2_000);
  }
</script>

<div class="truncate flex flex-col">
  <Tooltip distance={6} location="top">
    <button on:click={onCopy} class="truncate text-start" title={name}>
      <span class="source-code text-sm text-gray-800 font-medium truncate">
        {name}
      </span>
    </button>

    <TooltipContent slot="tooltip-content">
      {copied ? "Copied!" : "Click to copy"}
    </TooltipContent>
  </Tooltip>

  <span class="text-xs text-gray-500 font-normal truncate">
    {environmentLabel}
  </span>
</div>

<style lang="postcss">
  .source-code {
    font-family: "Source Code Variable", monospace;
  }
</style>
