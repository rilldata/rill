<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import type { HTTPError } from "@rilldata/web-common/lib/errors";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { useQueryClient } from "@tanstack/svelte-query";
  import ExpirationSelect from "./ExpirationSelect.svelte";
  import {
    ACCOUNT_TOKENS_PATH,
    DEFAULT_TOKEN_EXPIRATION,
    issuePersonalAccessToken,
    TOKEN_NAME_MAX_LENGTH,
    type TokenExpiration,
  } from "./personal-access-tokens";

  let {
    project,
    issuedToken = $bindable(null),
  }: {
    project: string;
    issuedToken?: string | null;
  } = $props();

  const queryClient = useQueryClient();

  // svelte-ignore state_referenced_locally
  let name = $state(m.mcp_token_name_default({ project }));
  let expiration = $state<TokenExpiration>(DEFAULT_TOKEN_EXPIRATION);
  let customDate = $state("");
  let error = $state<string | null>(null);
  let issuing = $state(false);

  let trimmedName = $derived(name.trim());
  let nameError = $derived(
    trimmedName.length > TOKEN_NAME_MAX_LENGTH
      ? m.access_tokens_name_too_long({ max: TOKEN_NAME_MAX_LENGTH })
      : "",
  );

  async function issueToken() {
    error = null;
    issuedToken = null;
    issuing = true;
    try {
      const { token } = await issuePersonalAccessToken(
        queryClient,
        trimmedName,
        expiration,
        customDate,
      );
      issuedToken = token || null;
    } catch (e) {
      error = (e as HTTPError).response?.data?.message || m.mcp_token_failed();
    } finally {
      issuing = false;
    }
  }
</script>

<div class="flex flex-col gap-y-3">
  <h4 class="text-sm font-medium text-fg-primary">
    {m.mcp_create_token_title()}
  </h4>
  <p class="text-sm text-fg-secondary">
    {@html m.mcp_create_token_desc({
      privateLabel: '<span class="font-medium">' + m.mcp_private() + "</span>",
      tokenLabel:
        '<span class="font-medium">' +
        m.mcp_personal_access_token() +
        "</span>",
    })}
    <a href={ACCOUNT_TOKENS_PATH}>{m.mcp_manage_tokens()}</a>
  </p>

  <Input
    id="mcp-token-name"
    label={m.common_name()}
    bind:value={name}
    errors={nameError}
    alwaysShowError
  />
  <ExpirationSelect id="mcp-token-expiration" bind:expiration bind:customDate />

  <div>
    <Button
      type="primary"
      onClick={issueToken}
      disabled={issuing || !trimmedName || !!nameError}
    >
      {issuing ? m.mcp_issuing() : m.mcp_create_token()}
    </Button>
  </div>

  {#if issuedToken}
    <div class="text-green-700 text-sm font-semibold">
      {m.mcp_token_created()}
    </div>
  {/if}

  {#if error}
    <div class="text-red-600 text-sm">{error}</div>
  {/if}
</div>
