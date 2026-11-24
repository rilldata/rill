<script lang="ts">
  import {
    createAdminServiceIssueUserAuthToken,
    getAdminServiceListUserAuthTokensQueryKey,
  } from "@rilldata/web-admin/client";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import type { HTTPError } from "@rilldata/web-common/lib/errors";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { DateTime } from "luxon";
  import ExpirationHint from "./ExpirationHint.svelte";
  import ExpirationSelect from "./ExpirationSelect.svelte";
  import {
    ACCOUNT_TOKENS_PATH,
    DEFAULT_TOKEN_EXPIRATION,
    getTokenExpiryDate,
    getTokenTTLMinutes,
    PERSONAL_ACCESS_TOKEN_CLIENT_ID,
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
  const issueTokenMutation = createAdminServiceIssueUserAuthToken();

  // svelte-ignore state_referenced_locally
  let name = $state(m.mcp_token_name_default({ project }));
  let expiration = $state<TokenExpiration>(DEFAULT_TOKEN_EXPIRATION);
  let customDate = $state<DateTime | null>(null);
  let error = $state<string | null>(null);

  let trimmedName = $derived(name.trim());
  let nameError = $derived(
    trimmedName.length > TOKEN_NAME_MAX_LENGTH
      ? m.access_tokens_name_too_long({ max: TOKEN_NAME_MAX_LENGTH })
      : "",
  );
  let expiresOn = $derived(getTokenExpiryDate(expiration, customDate));

  async function issueToken() {
    error = null;
    issuedToken = null;
    const now = DateTime.now();
    try {
      const resp = await $issueTokenMutation.mutateAsync({
        userId: "current",
        data: {
          displayName: trimmedName,
          clientId: PERSONAL_ACCESS_TOKEN_CLIENT_ID,
          ttlMinutes: String(
            getTokenTTLMinutes(
              getTokenExpiryDate(expiration, customDate, now),
              now,
            ),
          ),
        },
      });
      issuedToken = resp.token ?? null;
      void queryClient.invalidateQueries({
        queryKey: getAdminServiceListUserAuthTokensQueryKey("current"),
      });
    } catch (e) {
      error = (e as HTTPError).response?.data?.message || m.mcp_token_failed();
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

  <div class="grid grid-cols-[minmax(0,1fr)_180px] gap-x-3 items-start">
    <Input
      id="mcp-token-name"
      label={m.common_name()}
      labelGap={2}
      bind:value={name}
      errors={nameError}
      alwaysShowError
    />
    <ExpirationSelect
      id="mcp-token-expiration"
      bind:expiration
      bind:customDate
    />
  </div>
  <ExpirationHint {expiresOn} />

  <div>
    <Button
      type="primary"
      onClick={issueToken}
      disabled={$issueTokenMutation.isPending || !trimmedName || !!nameError}
    >
      {$issueTokenMutation.isPending ? m.mcp_issuing() : m.mcp_create_token()}
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
