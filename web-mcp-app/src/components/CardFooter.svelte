<script lang="ts">
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { ExternalLink, Maximize2, Minimize2 } from "lucide-svelte";

  type Secondary = "fullscreen" | "exit-fullscreen";

  let {
    openUrl,
    onOpenLink,
    secondary,
    onSecondary,
  }: {
    openUrl: string | undefined;
    onOpenLink: () => void;
    /** At most one more action next to "Open in Rill"; cards never show menus. */
    secondary?: Secondary;
    onSecondary?: () => void;
  } = $props();
</script>

{#if openUrl || secondary}
  <footer class="card-footer">
    {#if secondary === "fullscreen"}
      <button type="button" class="action" onclick={onSecondary}>
        <Maximize2 size="14" aria-hidden="true" />
        <span>{m.mcp_chart_fullscreen()}</span>
      </button>
    {:else if secondary === "exit-fullscreen"}
      <button type="button" class="action" onclick={onSecondary}>
        <Minimize2 size="14" aria-hidden="true" />
        <span>{m.mcp_chart_exit_fullscreen()}</span>
      </button>
    {/if}
    {#if openUrl}
      <button type="button" class="action primary" onclick={onOpenLink}>
        <ExternalLink size="14" aria-hidden="true" />
        <span>{m.mcp_chart_open_in_rill()}</span>
      </button>
    {/if}
  </footer>
{/if}

<style lang="postcss">
  .card-footer {
    @apply flex flex-wrap items-center justify-end gap-2 px-3 pb-3 pt-1;
  }
  .action {
    /* 44px targets per the host design guidelines. */
    @apply inline-flex items-center gap-1.5 min-h-[44px] px-3.5 text-[13px] font-medium;
    @apply rounded-[var(--radius)] border border-border bg-transparent text-fg-primary;
    @apply hover:bg-surface-hover active:bg-surface-active transition-colors cursor-pointer;
  }
  .action.primary {
    @apply text-fg-accent;
  }
</style>
