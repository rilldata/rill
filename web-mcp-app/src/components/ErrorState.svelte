<script lang="ts">
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { AlertCircle } from "lucide-svelte";
  import CardFooter from "./CardFooter.svelte";

  let {
    message,
    openUrl,
    onOpenLink,
  }: {
    message: string;
    openUrl: string | undefined;
    onOpenLink: () => void;
  } = $props();
</script>

<section class="state" role="alert">
  <div class="body">
    <AlertCircle size="22" aria-hidden="true" class="icon" />
    <h2 class="title">{m.mcp_chart_error_title()}</h2>
    <p class="text">{message}</p>
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
    @apply m-0 max-w-sm break-words text-xs leading-4 text-fg-tertiary;
  }
  .body :global(.icon) {
    @apply text-icon-muted;
  }
</style>
