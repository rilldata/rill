<script lang="ts">
  import Callout from "@rilldata/web-common/components/callout/Callout.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import type { DateTime } from "luxon";
  import { formatTokenDate } from "./personal-access-tokens";

  let { expiresOn }: { expiresOn: DateTime | null } = $props();
</script>

{#if expiresOn}
  <p class="text-xs text-fg-secondary">
    {m.access_tokens_expires_on({ date: formatTokenDate(expiresOn) })}
  </p>
{:else}
  <Callout level="warning">
    <span class="text-xs">{m.access_tokens_no_expiration_warning()}</span>
  </Callout>
{/if}
