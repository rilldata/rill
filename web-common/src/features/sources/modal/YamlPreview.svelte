<script lang="ts">
  import CopyIcon from "@rilldata/web-common/components/icons/CopyIcon.svelte";
  import Check from "@rilldata/web-common/components/icons/Check.svelte";
  import { InfoIcon } from "lucide-svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import type { V1ConnectorDriver } from "@rilldata/web-common/runtime-client";

  export let title: string;
  export let yaml: string;
  export let showAdditionalInfo: boolean = false;
  export let connector: V1ConnectorDriver | undefined = undefined;

  let copied = false;

  function copyYaml() {
    navigator.clipboard.writeText(yaml);
    copied = true;
    setTimeout(() => (copied = false), 2_000);
  }
</script>

<div>
  <div class="text-sm leading-none font-medium mb-4 flex items-center gap-1">
    {title}
    <slot name="title-action" />
  </div>
  <div class="relative">
    <button
      class="absolute top-2 right-2 p-1 rounded"
      type="button"
      aria-label="Copy YAML"
      on:click={copyYaml}
    >
      {#if copied}
        <Check size="16px" />
      {:else}
        <CopyIcon size="16px" />
      {/if}
    </button>
    <pre
      class="bg-muted p-3 rounded text-xs border border-gray-200 overflow-x-auto">{yaml}</pre>
  </div>
  {#if showAdditionalInfo && connector}
    <div class="mt-4 flex items-center gap-1">
      <span class="text-sm leading-none font-medium">Additional Information</span>
      <Tooltip location="right" alignment="middle" distance={8}>
        <div class="text-gray-500">
          <InfoIcon size="13px" />
        </div>
        <TooltipContent maxWidth="240px" slot="tooltip-content">
          External {connector.displayName} files are meant for local development
          only. They may run fine on your machine, but aren't reliably supported
          in production deployments—especially if the file is large (100MB) or outside
          the data directory.
        </TooltipContent>
      </Tooltip>
    </div>
  {/if}
  <slot />
  <!-- support need help, errors etc -->
</div>
