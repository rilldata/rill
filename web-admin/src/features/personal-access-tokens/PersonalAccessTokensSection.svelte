<script lang="ts">
  import { createAdminServiceIssueUserAuthToken } from "@rilldata/web-admin/client";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

  export let issuedToken: string | null = null;

  let error: string | null = null;
  let issuing = false;

  const issueTokenMutation = createAdminServiceIssueUserAuthToken();
  const manualClientId = "12345678-0000-0000-0000-000000000005"; // This comes from admin/database/database.go

  async function issueToken() {
    issuing = true;
    error = null;
    issuedToken = null;
    try {
      const resp = await $issueTokenMutation.mutateAsync({
        userId: "current",
        data: {
          displayName: "MCP Token",
          clientId: manualClientId,
          ttlMinutes: "0",
        },
      });
      issuedToken = resp.token;
    } catch (e) {
      error = e?.message || m.mcp_token_failed();
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
  </p>
  <div>
    <Button type="primary" onClick={issueToken} disabled={issuing}>
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
