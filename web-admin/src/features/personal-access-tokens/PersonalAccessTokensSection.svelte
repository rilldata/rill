<script lang="ts">
  import { createAdminServiceIssueUserAuthToken } from "@rilldata/web-admin/client";
  import Button from "@rilldata/web-common/components/button/Button.svelte";

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
      error = e?.message || "Failed to issue token. Please try again.";
    } finally {
      issuing = false;
    }
  }
</script>

<div class="flex flex-col gap-y-3">
  <h3 class="text-sm font-bold uppercase tracking-wide text-gray-900">
    Create a Personal Access Token
  </h3>
  <p class="text-sm text-gray-600">
    Because this project is <span class="font-medium">private</span>, you need a
    <span class="font-medium">personal access token</span> to use in your MCP configuration.
    This token authenticates your requests.
  </p>
  <div>
    <Button type="primary" onClick={issueToken} disabled={issuing}>
      {issuing ? "Issuing..." : "Create token"}
    </Button>
  </div>

  {#if issuedToken}
    <div class="text-green-700 text-sm font-semibold">
      Token created! Your new token is now included in the configuration snippet
      below.
    </div>
  {/if}

  {#if error}
    <div class="text-red-600 text-sm">{error}</div>
  {/if}
</div>
