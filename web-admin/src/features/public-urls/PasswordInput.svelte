<script lang="ts">
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import GeneratePasswordButton from "./GeneratePasswordButton.svelte";
  import { isValidPublicURLPassword } from "./password-policy";

  let {
    value = $bindable(""),
    id,
    placeholder = m.public_url_password_placeholder(),
    disabled = false,
  }: {
    value?: string;
    id: string;
    placeholder?: string;
    disabled?: boolean;
  } = $props();

  let invalid = $derived(value !== "" && !isValidPublicURLPassword(value));
</script>

<div class="flex w-full flex-col items-start gap-y-1">
  <Input
    {id}
    bind:value
    secret
    {placeholder}
    {disabled}
    errors={invalid ? m.public_url_password_length() : null}
    alwaysShowError
  />
  <GeneratePasswordButton
    {disabled}
    onGenerate={(password) => {
      value = password;
    }}
  />
</div>
