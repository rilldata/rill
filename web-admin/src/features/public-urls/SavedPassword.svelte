<script lang="ts">
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import PasswordCopyButton from "./PasswordCopyButton.svelte";

  let { password }: { password: string } = $props();
</script>

<div class="flex flex-col gap-y-2">
  <p class="text-xs text-fg-secondary">{m.public_url_password_save_notice()}</p>
  <div
    class="flex items-center justify-between gap-x-2 rounded border bg-surface-secondary px-2 py-1"
  >
    <code class="truncate font-mono text-xs" data-public-url-password={password}
      >{password}</code
    >
    <PasswordCopyButton getPassword={() => password} />
  </div>
</div>
