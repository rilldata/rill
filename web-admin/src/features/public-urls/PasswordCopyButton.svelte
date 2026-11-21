<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import Check from "@rilldata/web-common/components/icons/Check.svelte";
  import { tryCopyToClipboard } from "@rilldata/web-common/lib/actions/copy-to-clipboard";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { onDestroy } from "svelte";

  let {
    getPassword,
    label = m.public_url_copy_password_button(),
    disabled = false,
  }: {
    getPassword: () => string;
    label?: string;
    disabled?: boolean;
  } = $props();

  let copied = $state(false);
  let timeout: ReturnType<typeof setTimeout> | undefined;

  async function copy() {
    clearTimeout(timeout);
    copied = false;
    copied = await tryCopyToClipboard(getPassword());
    if (copied) timeout = setTimeout(() => (copied = false), 2_000);
  }

  onDestroy(() => clearTimeout(timeout));
</script>

<Button type="text" compact {disabled} onClick={copy}>
  {#if copied}
    <Check size="14px" />
    {m.public_url_copied_password_button()}
  {:else}
    {label}
  {/if}
</Button>
