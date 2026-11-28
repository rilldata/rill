<script lang="ts">
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { formatTokenDate, isExpiringSoon } from "../personal-access-tokens";

  let { expiresOn }: { expiresOn: string | undefined } = $props();

  let soon = $derived(isExpiringSoon(expiresOn));
</script>

{#if expiresOn}
  <div class="flex flex-col whitespace-nowrap" class:text-yellow-700={soon}>
    <span>{formatTokenDate(expiresOn)}</span>
    {#if soon}
      <span class="text-xs">{m.access_tokens_expires_soon()}</span>
    {/if}
  </div>
{:else}
  <span class="text-fg-secondary">{m.access_tokens_never_expires()}</span>
{/if}
