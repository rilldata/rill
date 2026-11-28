<script lang="ts">
  import CancelCircle from "@rilldata/web-common/components/icons/CancelCircle.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import MetadataValue from "./MetadataValue.svelte";

  /**
   * The dashboard of a report or an alert, linked with the saved state applied.
   * `missing`: the dashboard was deleted, so there is nothing to link to.
   * `stateLoadFailed`: the saved state could not be mapped to a url, so `href` opens the dashboard without it.
   */
  let {
    title,
    href,
    state,
    notExistMessage,
    stateLoadFailedMessage,
    openWithoutStateLabel,
  }: {
    title: string;
    href: string | undefined;
    state: "missing" | "stateLoadFailed" | "ok";
    notExistMessage: string;
    stateLoadFailedMessage: string;
    openWithoutStateLabel: string;
  } = $props();
</script>

<MetadataValue>
  {#if state === "missing"}
    <div class="flex items-center gap-x-1">
      {title}
      <Tooltip distance={8}>
        <CancelCircle size="16px" className="text-red-500" />
        <TooltipContent slot="tooltip-content">
          {notExistMessage}
        </TooltipContent>
      </Tooltip>
    </div>
  {:else if state === "stateLoadFailed"}
    <div class="flex flex-col gap-y-1">
      {title}
      <span class="text-xs text-fg-secondary">
        {stateLoadFailedMessage}
      </span>
      <a {href}>
        {openWithoutStateLabel}
      </a>
    </div>
  {:else if href}
    <a {href}>
      {title}
    </a>
  {:else}
    {title}
  {/if}
</MetadataValue>
