<script lang="ts">
  import Code from "@rilldata/web-common/components/icons/Code.svelte";
  import InfoCircle from "@rilldata/web-common/components/icons/InfoCircle.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import ProjectClone from "./ProjectClone.svelte";

  export let organization: string;
  export let project: string;
</script>

<div class="info-cell">
  <div class="cell-header">
    <span class="icon-muted"><Code size="16px" /></span>
    <span class="cell-label">Local Development</span>
    <Tooltip distance={8}>
      <a
        href="https://docs.rilldata.com/developers/guides/clone-a-project"
        target="_blank"
        rel="noreferrer noopener"
        class="info-link"
      >
        <InfoCircle size="14px" color="#9ca3af" />
      </a>
      <TooltipContent slot="tooltip-content">
        <p class="tooltip-text">Clone this project to develop locally.</p>
      </TooltipContent>
    </Tooltip>
  </div>
  <div class="cell-content">
    <div class="button-group">
      <ProjectClone {organization} {project} />
    </div>
  </div>
</div>

<style lang="postcss">
  .info-cell {
    @apply flex flex-col gap-1.5;
  }

  .cell-header {
    @apply flex items-center gap-1.5;
  }

  .cell-label {
    @apply text-xs font-medium text-fg-secondary uppercase tracking-wide;
  }

  .cell-content {
    @apply flex flex-col gap-1;
  }

  .button-group {
    @apply flex gap-2;
  }

  .icon-muted {
    @apply text-fg-secondary;
  }

  .icon-muted :global(svg path) {
    fill: currentColor;
  }

  .info-link {
    @apply flex items-center;
  }

  .info-link:hover :global(svg) {
    color: #6b7280;
  }

  .tooltip-text {
    @apply text-sm max-w-[200px];
  }
</style>
