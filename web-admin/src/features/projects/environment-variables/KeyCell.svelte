<script lang="ts">
  import { copyToClipboard } from "@rilldata/web-common/lib/actions/copy-to-clipboard";
  import { EnvironmentType } from "./types";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

  export let environment: string;
  export let name: string;

  // NOTE: if environment is empty, the variable is shared for all environments
  function getEnvironmentType(environment: string) {
    if (environment === EnvironmentType.UNDEFINED) {
      return m.env_development_production_label();
    }
    if (environment === EnvironmentType.DEVELOPMENT) {
      return m.env_development_label();
    }
    if (environment === EnvironmentType.PRODUCTION) {
      return m.env_production_label();
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
    <button onclick={onCopy} class="truncate text-start" title={name}>
      <span class="source-code text-sm text-fg-primary font-medium truncate">
        {name}
      </span>
    </button>

    <TooltipContent slot="tooltip-content">
      {copied ? m.env_copied_tooltip() : m.env_click_to_copy_tooltip()}
    </TooltipContent>
  </Tooltip>

  <span class="text-xs text-fg-secondary font-normal truncate">
    {environmentLabel}
  </span>
</div>

<style lang="postcss">
  .source-code {
    font-family: "Source Code Variable", monospace;
  }
</style>
