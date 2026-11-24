<script lang="ts">
  import {
    createAdminServiceIssueUserAuthToken,
    getAdminServiceListUserAuthTokensQueryKey,
  } from "@rilldata/web-admin/client";
  import { Button } from "@rilldata/web-common/components/button";
  import * as Dialog from "@rilldata/web-common/components/dialog";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import { tryCopyToClipboard } from "@rilldata/web-common/lib/actions/copy-to-clipboard";
  import type { HTTPError } from "@rilldata/web-common/lib/errors";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { DateTime } from "luxon";
  import ExpirationHint from "./ExpirationHint.svelte";
  import ExpirationSelect from "./ExpirationSelect.svelte";
  import {
    DEFAULT_TOKEN_EXPIRATION,
    formatTokenDate,
    getTokenExpiryDate,
    getTokenTTLMinutes,
    PERSONAL_ACCESS_TOKEN_CLIENT_ID,
    TOKEN_NAME_MAX_LENGTH,
    type TokenExpiration,
  } from "./personal-access-tokens";

  // Rendered inside the dialog's content, so every opening starts from an empty form.
  let { onDone }: { onDone: () => void } = $props();

  const formId = "create-access-token-form";

  const queryClient = useQueryClient();
  const issueToken = createAdminServiceIssueUserAuthToken();

  let name = $state("");
  let expiration = $state<TokenExpiration>(DEFAULT_TOKEN_EXPIRATION);
  let customDate = $state<DateTime | null>(null);
  let submitted = $state(false);
  let submitting = $state(false);
  let apiError = $state("");
  let copied = $state(false);
  // The secret is only in the issue response, so it is shown once and kept out of the query cache.
  let issued = $state<{
    token: string;
    name: string;
    expiresOn: DateTime | null;
  } | null>(null);

  let trimmedName = $derived(name.trim());
  let nameError = $derived(
    trimmedName.length > TOKEN_NAME_MAX_LENGTH
      ? m.access_tokens_name_too_long({ max: TOKEN_NAME_MAX_LENGTH })
      : submitted && !trimmedName
        ? m.access_tokens_name_required()
        : "",
  );
  let expiresOn = $derived(getTokenExpiryDate(expiration, customDate));

  async function submit() {
    submitted = true;
    if (!trimmedName || nameError || submitting) return;

    submitting = true;
    apiError = "";
    const now = DateTime.now();
    const tokenExpiresOn = getTokenExpiryDate(expiration, customDate, now);
    try {
      const resp = await $issueToken.mutateAsync({
        userId: "current",
        data: {
          clientId: PERSONAL_ACCESS_TOKEN_CLIENT_ID,
          displayName: trimmedName,
          ttlMinutes: String(getTokenTTLMinutes(tokenExpiresOn, now)),
        },
      });
      issued = {
        token: resp.token ?? "",
        name: trimmedName,
        expiresOn: tokenExpiresOn,
      };
      void queryClient.invalidateQueries({
        queryKey: getAdminServiceListUserAuthTokensQueryKey("current"),
      });
    } catch (e) {
      const error = e as HTTPError;
      apiError =
        error.response?.data?.message ?? m.access_tokens_create_failed();
    } finally {
      submitting = false;
    }
  }

  async function copyToken() {
    if (!issued) return;
    copied = await tryCopyToClipboard(issued.token);
    setTimeout(() => (copied = false), 2_000);
  }
</script>

{#if !issued}
  <Dialog.Header>
    <Dialog.Title>{m.access_tokens_create_title()}</Dialog.Title>
    <Dialog.Description>
      {m.access_tokens_create_description()}
    </Dialog.Description>
  </Dialog.Header>

  <form
    id={formId}
    class="flex flex-col gap-y-4"
    onsubmit={(e) => {
      e.preventDefault();
      void submit();
    }}
  >
    <Input
      id="access-token-name"
      label={m.common_name()}
      placeholder={m.access_tokens_name_placeholder()}
      bind:value={name}
      errors={nameError}
      alwaysShowError
      claimFocusOnMount
      onEnter={() => void submit()}
    />
    <div class="flex flex-col gap-y-2">
      <ExpirationSelect
        id="access-token-expiration"
        bind:expiration
        bind:customDate
      />
      <ExpirationHint {expiresOn} />
    </div>
    {#if apiError}
      <p class="text-destructive text-sm">{apiError}</p>
    {/if}
  </form>

  <Dialog.Footer>
    <Button type="secondary" onClick={onDone}>{m.common_cancel()}</Button>
    <Button
      type="primary"
      submitForm
      form={formId}
      loading={submitting}
      loadingCopy={m.access_tokens_creating()}
      disabled={submitting}
    >
      {m.access_tokens_create_submit()}
    </Button>
  </Dialog.Footer>
{:else}
  <Dialog.Header>
    <Dialog.Title>{m.access_tokens_copy_title()}</Dialog.Title>
    <Dialog.Description>
      {m.access_tokens_copy_description()}
    </Dialog.Description>
  </Dialog.Header>

  <div class="flex flex-col gap-y-2">
    <div class="flex items-start gap-x-2">
      <!-- Shown in full rather than in an input, which scrolls to its end when focused. -->
      <code
        class="flex-1 min-w-0 select-all break-all rounded-[2px] border bg-surface-background px-2 py-1.5 font-mono text-xs text-fg-primary"
      >
        {issued.token}
      </code>
      <Button type="secondary" onClick={copyToken}>
        {copied ? m.access_tokens_copied() : m.access_tokens_copy()}
      </Button>
    </div>
    <p class="text-xs text-fg-secondary">
      {issued.expiresOn
        ? m.access_tokens_created_expires({
            name: issued.name,
            date: formatTokenDate(issued.expiresOn),
          })
        : m.access_tokens_created_never_expires({ name: issued.name })}
      {m.access_tokens_usage_hint()}
    </p>
  </div>

  <Dialog.Footer>
    <Button type="primary" onClick={onDone}>{m.access_tokens_done()}</Button>
  </Dialog.Footer>
{/if}
