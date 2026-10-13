<!--
  Fills a password field with a generated password and copies it to the clipboard,
  so the creator can paste it wherever they share it (the field itself shows it masked).
-->
<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import Check from "@rilldata/web-common/components/icons/Check.svelte";
  import { copyToClipboard } from "@rilldata/web-common/lib/actions/copy-to-clipboard";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { generatePassword } from "./generate-password";

  let {
    onGenerate,
    disabled = false,
  }: { onGenerate: (password: string) => void; disabled?: boolean } = $props();

  let copied = $state(false);
  let copiedTimeout: ReturnType<typeof setTimeout> | undefined;

  function generateAndCopy() {
    const password = generatePassword();
    onGenerate(password);
    copyToClipboard(password, undefined, false);
    copied = true;
    clearTimeout(copiedTimeout);
    copiedTimeout = setTimeout(() => {
      copied = false;
    }, 2_000);
  }
</script>

<Button type="text" compact {disabled} onClick={generateAndCopy}>
  {#if copied}
    <Check size="14px" />
    {m.public_url_copied_password_button()}
  {:else}
    {m.public_url_generate_copy_password_button()}
  {/if}
</Button>
