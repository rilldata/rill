<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import { slide, fly } from "svelte/transition";
  import { X } from "lucide-svelte";
  import {
    createGetServiceTokenQuery,
    createGetUserTokenQuery,
  } from "./token-queries";
  import DeleteTokenDialog from "./DeleteTokenDialog.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import { formatDistanceToNow, format } from "date-fns";

  export let tokenId: string;
  export let tokenType: "service" | "user";
  export let orgId: string;
  export let open: boolean;

  const dispatch = createEventDispatcher<{
    close: void;
    deleted: void;
  }>();

  let showDeleteDialog = false;

  // Conditionally create the appropriate query based on token type
  $: serviceTokenQuery =
    tokenType === "service" && open && tokenId
      ? createGetServiceTokenQuery(orgId, tokenId)
      : undefined;

  $: userTokenQuery =
    tokenType === "user" && open && tokenId
      ? createGetUserTokenQuery(tokenId)
      : undefined;

  $: activeQuery =
    tokenType === "service" ? serviceTokenQuery : userTokenQuery;
  $: token = $activeQuery?.data ?? undefined;
  $: isLoading = $activeQuery?.isLoading ?? false;
  $: isError = $activeQuery?.isError ?? false;
  $: error = $activeQuery?.error;

  function handleClose() {
    dispatch("close");
  }

  function handleKeydown(event: KeyboardEvent) {
    if (event.key === "Escape" && open) {
      handleClose();
    }
  }

  function handleOverlayClick() {
    handleClose();
  }

  function handleRevokeClick() {
    showDeleteDialog = true;
  }

  async function handleDeleteConfirm() {
    showDeleteDialog = false;
    dispatch("deleted");
    dispatch("close");
  }

  function formatAbsoluteDate(dateStr: string | undefined | null): string {
    if (!dateStr) return "—";
    try {
      return format(new Date(dateStr), "MMM d, yyyy 'at' h:mm a");
    } catch {
      return "—";
    }
  }

  function formatRelativeDate(dateStr: string | undefined | null): string {
    if (!dateStr) return "";
    try {
      return formatDistanceToNow(new Date(dateStr), { addSuffix: true });
    } catch {
      return "";
    }
  }

  function getPermissionLabel(permissions: string | undefined | null): string {
    if (!permissions) return "—";
    const lower = permissions.toLowerCase();
    if (lower.includes("admin")) return "Admin";
    if (lower.includes("editor") || lower.includes("write")) return "Editor";
    if (lower.includes("read") || lower.includes("viewer")) return "Read";
    return permissions;
  }

  function getPermissionColor(
    permissions: string | undefined | null,
  ): string {
    if (!permissions) return "bg-gray-100 text-gray-700";
    const lower = permissions.toLowerCase();
    if (lower.includes("admin")) return "bg-red-100 text-red-800";
    if (lower.includes("editor") || lower.includes("write"))
      return "bg-yellow-100 text-yellow-800";
    return "bg-blue-100 text-blue-800";
  }

  // Extract fields that may vary between service and user token shapes
  $: tokenName = token?.name ?? token?.displayName ?? "";
  $: tokenDescription = token?.description ?? "";
  $: tokenPrefix = token?.tokenPrefix ?? token?.prefix ?? "";
  $: createdAt = token?.createdAt ?? token?.createdOn ?? null;
  $: lastUsedAt = token?.lastUsedAt ?? token?.lastUsedOn ?? null;
  $: expiresAt = token?.expiresAt ?? token?.expiresOn ?? null;
  $: permissions = token?.permissions ?? token?.role ?? null;
  $: scope = token?.projectName
    ? `Project: ${token.projectName}`
    : token?.organizationName
      ? `Organization: ${token.organizationName}`
      : "Organization";
  $: createdBy = token?.createdByEmail ?? token?.createdBy ?? null;
  $: lastUsedIp =
    token?.lastUsedIpAddress ?? token?.lastUsedIp ?? null;
</script>

<svelte:window on:keydown={handleKeydown} />

{#if open}
  <!-- Overlay backdrop -->
  <div
    class="fixed inset-0 z-40 bg-black/30 transition-opacity"
    on:click={handleOverlayClick}
    on:keydown={(e) => e.key === "Enter" && handleOverlayClick()}
    role="button"
    tabindex="-1"
    aria-label="Close drawer"
  />

  <!-- Drawer panel -->
  <aside
    class="fixed right-0 top-0 z-50 flex h-full w-full max-w-md flex-col border-l border-gray-200 bg-white shadow-xl"
    transition:fly={{ x: 400, duration: 250, opacity: 1 }}
    role="dialog"
    aria-modal="true"
    aria-label="Token details"
  >
    <!-- Header -->
    <div
      class="flex items-center justify-between border-b border-gray-200 px-6 py-4"
    >
      <div class="flex items-center gap-2">
        <h2 class="text-lg font-semibold text-gray-900">Token Details</h2>
        <span
          class="rounded-full px-2 py-0.5 text-xs font-medium {tokenType ===
          'service'
            ? 'bg-purple-100 text-purple-800'
            : 'bg-green-100 text-green-800'}"
        >
          {tokenType === "service" ? "Service" : "Personal"}
        </span>
      </div>
      <button
        on:click={handleClose}
        class="rounded-md p-1 text-gray-400 hover:bg-gray-100 hover:text-gray-600 focus:outline-none focus:ring-2 focus:ring-primary-500"
        aria-label="Close"
      >
        <X size={20} />
      </button>
    </div>

    <!-- Content -->
    <div class="flex-1 overflow-y-auto px-6 py-5">
      {#if isLoading}
        <!-- Loading skeleton -->
        <div class="animate-pulse space-y-6">
          <div>
            <div class="mb-3 h-4 w-20 rounded bg-gray-200" />
            <div class="space-y-3">
              <div class="h-4 w-3/4 rounded bg-gray-200" />
              <div class="h-4 w-1/2 rounded bg-gray-200" />
              <div class="h-4 w-2/3 rounded bg-gray-200" />
            </div>
          </div>
          <div>
            <div class="mb-3 h-4 w-24 rounded bg-gray-200" />
            <div class="space-y-3">
              <div class="h-4 w-1/2 rounded bg-gray-200" />
              <div class="h-4 w-2/3 rounded bg-gray-200" />
            </div>
          </div>
          <div>
            <div class="mb-3 h-4 w-28 rounded bg-gray-200" />
            <div class="space-y-3">
              <div class="h-4 w-3/4 rounded bg-gray-200" />
              <div class="h-4 w-1/2 rounded bg-gray-200" />
              <div class="h-4 w-2/3 rounded bg-gray-200" />
            </div>
          </div>
        </div>
      {:else if isError}
        <div class="rounded-lg border border-red-200 bg-red-50 p-4">
          <p class="text-sm font-medium text-red-800">
            Failed to load token details
          </p>
          <p class="mt-1 text-sm text-red-600">
            {error?.message ?? "An unexpected error occurred."}
          </p>
          <button
            class="mt-3 text-sm font-medium text-red-700 underline hover:text-red-900"
            on:click={() => $activeQuery?.refetch()}
          >
            Try again
          </button>
        </div>
      {:else if token}
        <div class="space-y-6">
          <!-- General Section -->
          <section>
            <h3
              class="mb-3 text-xs font-semibold uppercase tracking-wider text-gray-500"
            >
              General
            </h3>
            <dl class="space-y-3">
              <div>
                <dt class="text-sm font-medium text-gray-500">Name</dt>
                <dd class="mt-0.5 text-sm text-gray-900">
                  {tokenName || "—"}
                </dd>
              </div>
              {#if tokenDescription}
                <div>
                  <dt class="text-sm font-medium text-gray-500">
                    Description
                  </dt>
                  <dd class="mt-0.5 text-sm text-gray-900">
                    {tokenDescription}
                  </dd>
                </div>
              {/if}
              <div>
                <dt class="text-sm font-medium text-gray-500">Token Prefix</dt>
                <dd class="mt-0.5">
                  <code
                    class="rounded bg-gray-100 px-2 py-0.5 font-mono text-sm text-gray-800"
                  >
                    {tokenPrefix || "—"}
                  </code>
                </dd>
              </div>
              <div>
                <dt class="text-sm font-medium text-gray-500">Type</dt>
                <dd class="mt-0.5">
                  <span
                    class="inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium {tokenType ===
                    'service'
                      ? 'bg-purple-100 text-purple-800'
                      : 'bg-green-100 text-green-800'}"
                  >
                    {tokenType === "service"
                      ? "Service Token"
                      : "Personal Token"}
                  </span>
                </dd>
              </div>
            </dl>
          </section>

          <!-- Access Section (service tokens only) -->
          {#if tokenType === "service"}
            <section>
              <h3
                class="mb-3 text-xs font-semibold uppercase tracking-wider text-gray-500"
              >
                Access
              </h3>
              <dl class="space-y-3">
                <div>
                  <dt class="text-sm font-medium text-gray-500">Scope</dt>
                  <dd class="mt-0.5 text-sm text-gray-900">{scope}</dd>
                </div>
                <div>
                  <dt class="text-sm font-medium text-gray-500">
                    Permissions
                  </dt>
                  <dd class="mt-0.5">
                    <span
                      class="inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium {getPermissionColor(
                        permissions,
                      )}"
                    >
                      {getPermissionLabel(permissions)}
                    </span>
                  </dd>
                </div>
              </dl>
            </section>
          {/if}

          <!-- Timestamps Section -->
          <section>
            <h3
              class="mb-3 text-xs font-semibold uppercase tracking-wider text-gray-500"
            >
              Timestamps
            </h3>
            <dl class="space-y-3">
              <div>
                <dt class="text-sm font-medium text-gray-500">Created</dt>
                <dd class="mt-0.5 text-sm text-gray-900">
                  {#if createdAt}
                    <span>{formatAbsoluteDate(createdAt)}</span>
                    <span class="ml-1 text-gray-500"
                      >({formatRelativeDate(createdAt)})</span
                    >
                  {:else}
                    —
                  {/if}
                </dd>
              </div>
              <div>
                <dt class="text-sm font-medium text-gray-500">Last Used</dt>
                <dd class="mt-0.5 text-sm text-gray-900">
                  {#if lastUsedAt}
                    <span>{formatAbsoluteDate(lastUsedAt)}</span>
                    <span class="ml-1 text-gray-500"
                      >({formatRelativeDate(lastUsedAt)})</span
                    >
                  {:else}
                    <span class="text-gray-400">Never used</span>
                  {/if}
                </dd>
              </div>
              {#if tokenType === "user"}
                <div>
                  <dt class="text-sm font-medium text-gray-500">Expires</dt>
                  <dd class="mt-0.5 text-sm text-gray-900">
                    {#if expiresAt}
                      {@const expiresDate = new Date(expiresAt)}
                      {@const isExpired = expiresDate < new Date()}
                      <span class={isExpired ? "text-red-600" : ""}
                        >{formatAbsoluteDate(expiresAt)}</span
                      >
                      <span
                        class="ml-1 {isExpired
                          ? 'text-red-500'
                          : 'text-gray-500'}"
                        >({isExpired ? "Expired " : ""}{formatRelativeDate(
                          expiresAt,
                        )})</span
                      >
                      {#if isExpired}
                        <span
                          class="ml-1 inline-flex items-center rounded-full bg-red-100 px-2 py-0.5 text-xs font-medium text-red-700"
                        >
                          Expired
                        </span>
                      {/if}
                    {:else}
                      <span class="text-gray-400">Never</span>
                    {/if}
                  </dd>
                </div>
              {/if}
            </dl>
          </section>

          <!-- Additional Info Section -->
          {#if createdBy || lastUsedIp}
            <section>
              <h3
                class="mb-3 text-xs font-semibold uppercase tracking-wider text-gray-500"
              >
                Additional
              </h3>
              <dl class="space-y-3">
                {#if createdBy}
                  <div>
                    <dt class="text-sm font-medium text-gray-500">
                      Created By
                    </dt>
                    <dd class="mt-0.5 text-sm text-gray-900">{createdBy}</dd>
                  </div>
                {/if}
                {#if lastUsedIp}
                  <div>
                    <dt class="text-sm font-medium text-gray-500">
                      Last Used IP
                    </dt>
                    <dd class="mt-0.5">
                      <code
                        class="rounded bg-gray-100 px-2 py-0.5 font-mono text-sm text-gray-800"
                      >
                        {lastUsedIp}
                      </code>
                    </dd>
                  </div>
                {/if}
              </dl>
            </section>
          {/if}
        </div>
      {/if}
    </div>

    <!-- Footer with revoke button -->
    {#if token && !isLoading && !isError}
      <div class="border-t border-gray-200 px-6 py-4">
        <Button
          type="secondary"
          on:click={handleRevokeClick}
          label="Revoke Token"
          danger
          wide
        >
          Revoke Token
        </Button>
      </div>
    {/if}
  </aside>
{/if}

<!-- Delete confirmation dialog -->
{#if showDeleteDialog}
  <DeleteTokenDialog
    tokenName={tokenName}
    tokenType={tokenType}
    onConfirm={handleDeleteConfirm}
    on:close={() => (showDeleteDialog = false)}
  />
{/if}