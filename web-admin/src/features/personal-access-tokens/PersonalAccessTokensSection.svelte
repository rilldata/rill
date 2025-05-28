<script lang="ts">
  import { createAdminServiceIssueUserAuthToken } from "@rilldata/web-admin/client";
  import Button from "@rilldata/web-common/components/button/Button.svelte";

  let showIssueDialog = false;
  let issuedToken: string | null = null;
  let error: string | null = null;
  let issuing = false;

  const issueTokenMutation = createAdminServiceIssueUserAuthToken();
  const manualClientId = "12345678-0000-0000-0000-000000000005"; // This comes from admin/database/database.go

  async function issueToken() {
    issuing = true;
    error = null;
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
      showIssueDialog = false;
    } catch (e) {
      error = e?.message || "Failed to issue token. Please try again.";
    } finally {
      issuing = false;
    }
  }

  function openDialog() {
    issuedToken = null;
    showIssueDialog = true;
    error = null;
  }
</script>

<div class="mb-8">
  <h2 class="text-xl font-semibold mb-2">Personal Access Token</h2>
  <p class="mb-4 text-gray-600">
    You need a <span class="font-medium">personal access token</span> to use in your
    MCP configuration for private projects. This token authenticates your requests.
  </p>
  <Button type="primary" on:click={openDialog}>Create token</Button>

  {#if showIssueDialog}
    <div
      class="fixed inset-0 bg-black bg-opacity-30 flex items-center justify-center z-50"
    >
      <div class="bg-white rounded shadow-lg p-6 w-full max-w-md">
        <h3 class="text-lg font-semibold mb-2">Create Personal Access Token</h3>
        <p class="mb-4 text-gray-600 text-sm">
          This token will be shown only once. Store it securely.
        </p>
        {#if error}
          <div class="text-red-600 mb-2">{error}</div>
        {/if}
        <div class="flex gap-2 justify-end">
          <Button type="secondary" on:click={() => (showIssueDialog = false)}>
            Cancel
          </Button>
          <Button type="primary" on:click={issueToken} disabled={issuing}>
            {issuing ? "Issuing..." : "Create Token"}
          </Button>
        </div>
      </div>
    </div>
  {/if}

  {#if issuedToken}
    <div class="mt-6 p-4 bg-gray-100 rounded">
      <div class="mb-2 font-semibold text-gray-700">Your new token:</div>
      <div class="flex items-center gap-2 mb-2">
        <code class="bg-white px-2 py-1 rounded font-mono text-sm"
          >{issuedToken}</code
        >
        <Button
          type="secondary"
          on:click={() => {
            navigator.clipboard.writeText(issuedToken);
          }}>Copy</Button
        >
      </div>
      <div class="text-xs text-gray-500">
        This token is shown only once. Store it securely.
      </div>
    </div>
  {/if}
</div>
