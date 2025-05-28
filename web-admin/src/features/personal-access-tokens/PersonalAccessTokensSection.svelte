<script lang="ts">
  import { createAdminServiceIssueUserAuthToken } from "@rilldata/web-admin/client";
  import Button from "@rilldata/web-common/components/button/Button.svelte";

  let issuedToken: string | null = null;
  let error: string | null = null;
  let issuing = false;
  let copied = false;
  let copyTimeout: ReturnType<typeof setTimeout> | null = null;

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

  function handleCopy() {
    if (issuedToken) {
      navigator.clipboard.writeText(issuedToken);
      copied = true;
      if (copyTimeout) clearTimeout(copyTimeout);
      copyTimeout = setTimeout(() => {
        copied = false;
      }, 1500);
    }
  }
</script>

<div class="mb-2">
  <h2 class="text-xl font-semibold mb-2">Personal Access Token</h2>
  <p class="mb-4 text-gray-600">
    Because this project is <span class="font-medium">private</span>, you need a
    <span class="font-medium">personal access token</span> to use in your MCP configuration.
    This token authenticates your requests.
  </p>
  <Button type="primary" on:click={issueToken} disabled={issuing}>
    {issuing ? "Issuing..." : "Create token"}
  </Button>

  {#if error}
    <div class="text-red-600 mt-2">{error}</div>
  {/if}

  {#if issuedToken}
    <div class="mt-6 p-4 bg-gray-100 rounded">
      <div class="mb-2 font-semibold text-gray-700">Your new token:</div>
      <div class="flex items-center gap-2 mb-2">
        <code class="bg-white px-2 py-1 rounded font-mono text-sm"
          >{issuedToken}</code
        >
        <Button type="secondary" on:click={handleCopy}>
          {#if copied}Copied!{:else}Copy token{/if}</Button
        >
      </div>
      <div class="text-xs text-gray-500">
        This token is shown only once. Store it securely.
      </div>
    </div>
  {/if}
</div>
