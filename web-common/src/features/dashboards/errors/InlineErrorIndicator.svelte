<script lang="ts">
  import * as Tooltip from "@rilldata/web-common/components/tooltip-v2";
  import { copyToClipboard } from "@rilldata/web-common/lib/actions/copy-to-clipboard";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { AlertTriangleIcon, Copy } from "lucide-svelte";

  export let message: string | undefined;
  export let compact = false;
</script>

<Tooltip.Root>
  <Tooltip.Trigger>
    {#snippet child({ props })}
      <button
        {...props}
        type="button"
        class="inline-flex items-center gap-1 text-fg-secondary text-xs"
        aria-label={m.dashboard_error_occurred_hover()}
      >
        <AlertTriangleIcon class="text-red-500" size="16px" />
        {#if !compact}
          <span>{m.dashboard_error_occurred()}</span>
        {/if}
      </button>
    {/snippet}
  </Tooltip.Trigger>

  <Tooltip.Content side="top" sideOffset={6} class="max-w-md p-0">
    <div class="flex items-start gap-2 p-2">
      <div
        class="text-fg-inverse whitespace-pre-wrap break-words max-h-60 overflow-auto text-xs font-mono"
      >
        {message || m.dashboard_no_additional_details()}
      </div>
      {#if message}
        <button
          type="button"
          class="shrink-0 p-1 rounded hover:bg-white/10 transition-colors text-fg-inverse"
          on:click={() =>
            copyToClipboard(message, m.dashboard_copied_error_clipboard())}
          title={m.dashboard_copy_error_clipboard()}
          aria-label={m.dashboard_copy_error_clipboard()}
        >
          <Copy size="14px" />
        </button>
      {/if}
    </div>
  </Tooltip.Content>
</Tooltip.Root>
