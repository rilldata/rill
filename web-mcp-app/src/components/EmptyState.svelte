<script lang="ts">
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { BarChart3 } from "lucide-svelte";
  import CardFooter from "./CardFooter.svelte";

  let {
    openUrl,
    onOpenLink,
  }: { openUrl: string | undefined; onOpenLink: () => void } = $props();
</script>

<section class="state" role="status">
  <div class="body">
    <BarChart3 size="22" aria-hidden="true" class="icon" />
    <h2 class="title">{m.mcp_chart_empty_title()}</h2>
    <p class="text">{m.mcp_chart_empty_body()}</p>
  </div>
  <CardFooter {openUrl} {onOpenLink} />
</section>

<style lang="postcss">
  .state {
    @apply flex w-full flex-col;
  }
  .body {
    @apply flex flex-col items-center justify-center gap-1.5 px-6 pt-8 pb-4 text-center;
    min-height: 180px;
  }
  .title {
    @apply m-0 text-sm font-semibold text-fg-primary;
  }
  .text {
    @apply m-0 max-w-sm text-xs leading-4 text-fg-tertiary;
  }
  .body :global(.icon) {
    @apply text-icon-muted;
  }
</style>
