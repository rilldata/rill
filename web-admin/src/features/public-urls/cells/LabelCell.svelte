<script lang="ts">
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { ExternalLinkIcon, LockIcon } from "lucide-svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

  let {
    displayName,
    dashboardTitle,
    url,
    passwordProtected = false,
  }: {
    displayName: string;
    dashboardTitle: string;
    url: string;
    passwordProtected?: boolean;
  } = $props();
</script>

<div class="flex items-center gap-x-2 truncate">
  <a
    href={url}
    target="_blank"
    rel="noopener noreferrer"
    class="flex items-center gap-x-2 hover:text-accent-primary-action truncate"
  >
    <ExternalLinkIcon size={14} class="shrink-0" />
    <span class="truncate">
      {displayName || dashboardTitle || m.public_url_untitled()}
    </span>
  </a>
  {#if passwordProtected}
    <Tooltip distance={8}>
      <LockIcon
        size={12}
        class="shrink-0 text-fg-secondary"
        aria-label={m.public_url_password_protected()}
      />
      <TooltipContent slot="tooltip-content">
        {m.public_url_password_protected()}
      </TooltipContent>
    </Tooltip>
  {/if}
</div>
