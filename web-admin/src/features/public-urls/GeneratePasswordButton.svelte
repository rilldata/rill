<!--
  Fills a password field with a generated password and copies it to the clipboard,
  so the creator can paste it wherever they share it (the field itself shows it masked).
-->
<script lang="ts">
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { generatePassword } from "./generate-password";
  import PasswordCopyButton from "./PasswordCopyButton.svelte";

  let {
    onGenerate,
    disabled = false,
  }: { onGenerate: (password: string) => void; disabled?: boolean } = $props();

  function generate() {
    const password = generatePassword();
    onGenerate(password);
    return password;
  }
</script>

<PasswordCopyButton
  getPassword={generate}
  label={m.public_url_generate_copy_password_button()}
  {disabled}
/>
