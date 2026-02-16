<script lang="ts">
  import { page } from "$app/stores";
  import { goto } from "$app/navigation";
  import { derived } from "svelte/store";
  import ServiceTokenList from "@rilldata/web-admin/src/lib/features/tokens/ServiceTokenList.svelte";
  import UserTokenList from "@rilldata/web-admin/src/lib/features/tokens/UserTokenList.svelte";
  import CreateServiceTokenDialog from "@rilldata/web-admin/src/lib/features/tokens/CreateServiceTokenDialog.svelte";
  import CreateUserTokenDialog from "@rilldata/web-admin/src/lib/features/tokens/CreateUserTokenDialog.svelte";
  import { Button } from "@rilldata/web-common/src/components/button";
  import {
    createServiceTokenListQuery,
    createUserTokenListQuery,
  } from "@rilldata/web-admin/src/lib/features/tokens/token-queries";

  const TAB_SERVICE = "service";
  const TAB_USER = "user";
  type TabType = typeof TAB_SERVICE | typeof TAB_USER;

  $: organization = $page.params.organization;

  // Derive active tab from URL query param for deep-linking
  $: urlTab = $page.url.searchParams.get("tab");
  $: activeTab = (urlTab === TAB_USER ? TAB_USER : TAB_SERVICE) as TabType;

  // Dialog state
  let showCreateServiceDialog = false;
  let showCreateUserDialog = false;

  // Prefetch queries to detect loading/error at the page level
  $: serviceTokensQuery = createServiceTokenListQuery(organization);
  $: userTokensQuery = createUserTokenListQuery();

  // Derive the active query based on current tab for page-level loading/error
  $: activeQuery = activeTab === TAB_SERVICE ? serviceTokensQuery : userTokensQuery;

  // Page title
  $: pageTitle = `Tokens - Settings - ${organization}`;

  function setActiveTab(tab: TabType) {
    const url = new URL($page.url);
    if (tab === TAB_SERVICE) {
      url.searchParams.delete("tab");
    } else {
      url.searchParams.set("tab", tab);
    }
    goto(url.toString(), { replaceState: true, noScroll: true });
  }

  function handleCreateClick() {
    if (activeTab === TAB_USER) {
      showCreateUserDialog = true;
    } else {
      showCreateServiceDialog = true;
    }
  }

  function handleTabKeydown(event: KeyboardEvent, tab: TabType) {
    if (event.key === "Enter" || event.key === " ") {
      event.preventDefault();
      setActiveTab(tab);
    }
    // Arrow key navigation between tabs
    if (event.key === "ArrowRight" || event.key === "ArrowLeft") {
      event.preventDefault();
      const nextTab = tab === TAB_SERVICE ? TAB_USER : TAB_SERVICE;
      setActiveTab(nextTab);
      // Focus the other tab button
      const nextEl = document.querySelector(`[data-tab="${nextTab}"]`) as HTMLElement;
      nextEl?.focus();
    }
  }

  function handleRetry() {
    if (activeTab === TAB_SERVICE) {
      $serviceTokensQuery.refetch();
    } else {
      $userTokensQuery.refetch();
    }
  }

  function handleKeydown(event: KeyboardEvent) {
    if (event.key === "Escape") {
      showCreateServiceDialog = false;
      showCreateUserDialog = false;
    }
  }
</script>

<svelte:head>
  <title>{pageTitle}</title>
</svelte:head>

<svelte:window on:keydown={handleKeydown} />

<div class="tokens-page">
  <!-- Page Header -->
  <div class="tokens-header">
    <div class="tokens-header-text">
      <h1 class="tokens-title">Tokens</h1>
      <p class="tokens-description">
        Manage service tokens for CI/CD integrations and personal tokens for CLI and API access.
      </p>
    </div>
    <div class="tokens-header-actions">
      <Button on:click={handleCreateClick} type="primary">
        Create Token
      </Button>
    </div>
  </div>

  <!-- Tab Bar -->
  <div class="tokens-tabs" role="tablist" aria-label="Token type tabs">
    <button
      class="tokens-tab"
      class:active={activeTab === TAB_SERVICE}
      role="tab"
      aria-selected={activeTab === TAB_SERVICE}
      aria-controls="panel-service"
      tabindex={activeTab === TAB_SERVICE ? 0 : -1}
      data-tab={TAB_SERVICE}
      on:click={() => setActiveTab(TAB_SERVICE)}
      on:keydown={(e) => handleTabKeydown(e, TAB_SERVICE)}
    >
      Service Tokens
    </button>
    <button
      class="tokens-tab"
      class:active={activeTab === TAB_USER}
      role="tab"
      aria-selected={activeTab === TAB_USER}
      aria-controls="panel-user"
      tabindex={activeTab === TAB_USER ? 0 : -1}
      data-tab={TAB_USER}
      on:click={() => setActiveTab(TAB_USER)}
      on:keydown={(e) => handleTabKeydown(e, TAB_USER)}
    >
      User Tokens
    </button>
  </div>

  <!-- Tab Content -->
  <div class="tokens-content">
    {#if $activeQuery.isLoading}
      <!-- Page-level loading skeleton -->
      <div class="tokens-skeleton" aria-busy="true" aria-label="Loading tokens">
        <div class="skeleton-toolbar">
          <div class="skeleton-search skeleton-block" />
          <div class="skeleton-button skeleton-block" />
        </div>
        <div class="skeleton-table">
          <div class="skeleton-row skeleton-header-row">
            {#each Array(6) as _}
              <div class="skeleton-cell skeleton-block" />
            {/each}
          </div>
          {#each Array(5) as _, i}
            <div class="skeleton-row" style="animation-delay: {i * 75}ms">
              {#each Array(6) as _}
                <div class="skeleton-cell skeleton-block" />
              {/each}
            </div>
          {/each}
        </div>
      </div>
    {:else if $activeQuery.isError}
      <!-- Page-level error state -->
      <div class="tokens-error" role="alert">
        <div class="tokens-error-icon">
          <svg
            xmlns="http://www.w3.org/2000/svg"
            width="48"
            height="48"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="1.5"
            stroke-linecap="round"
            stroke-linejoin="round"
          >
            <circle cx="12" cy="12" r="10" />
            <line x1="12" y1="8" x2="12" y2="12" />
            <line x1="12" y1="16" x2="12.01" y2="16" />
          </svg>
        </div>
        <h2 class="tokens-error-title">Failed to load tokens</h2>
        <p class="tokens-error-message">
          {$activeQuery.error?.message ?? "An unexpected error occurred while loading your tokens. Please try again."}
        </p>
        <Button on:click={handleRetry} type="secondary">
          Retry
        </Button>
      </div>
    {:else}
      {#if activeTab === TAB_SERVICE}
        <div
          id="panel-service"
          role="tabpanel"
          aria-labelledby="tab-service"
          class="tokens-panel"
        >
          <ServiceTokenList orgId={organization} />
        </div>
      {:else}
        <div
          id="panel-user"
          role="tabpanel"
          aria-labelledby="tab-user"
          class="tokens-panel"
        >
          <UserTokenList />
        </div>
      {/if}
    {/if}
  </div>
</div>

<!-- Dialogs -->
{#if showCreateServiceDialog}
  <CreateServiceTokenDialog
    orgId={organization}
    on:close={() => (showCreateServiceDialog = false)}
  />
{/if}

{#if showCreateUserDialog}
  <CreateUserTokenDialog
    on:close={() => (showCreateUserDialog = false)}
  />
{/if}

<style>
  .tokens-page {
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
    width: 100%;
    max-width: 1200px;
    padding: 1.5rem;
  }

  /* Header */
  .tokens-header {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 1rem;
    flex-wrap: wrap;
  }

  .tokens-header-text {
    flex: 1;
    min-width: 200px;
  }

  .tokens-title {
    font-size: 1.5rem;
    font-weight: 600;
    color: var(--color-text-primary, #111827);
    margin: 0;
    line-height: 1.3;
  }

  .tokens-description {
    font-size: 0.875rem;
    color: var(--color-text-secondary, #6b7280);
    margin: 0.25rem 0 0;
    line-height: 1.5;
  }

  .tokens-header-actions {
    flex-shrink: 0;
  }

  /* Tabs */
  .tokens-tabs {
    display: flex;
    border-bottom: 1px solid var(--color-border, #e5e7eb);
    gap: 0;
  }

  .tokens-tab {
    padding: 0.625rem 1rem;
    font-size: 0.875rem;
    font-weight: 500;
    color: var(--color-text-secondary, #6b7280);
    background: none;
    border: none;
    border-bottom: 2px solid transparent;
    cursor: pointer;
    transition: color 0.15s, border-color 0.15s;
    white-space: nowrap;
    outline: none;
  }

  .tokens-tab:hover {
    color: var(--color-text-primary, #111827);
  }

  .tokens-tab:focus-visible {
    outline: 2px solid var(--color-primary, #6366f1);
    outline-offset: -2px;
    border-radius: 2px;
  }

  .tokens-tab.active {
    color: var(--color-primary, #6366f1);
    border-bottom-color: var(--color-primary, #6366f1);
  }

  /* Content */
  .tokens-content {
    min-height: 300px;
    overflow-x: auto;
    -webkit-overflow-scrolling: touch;
  }

  .tokens-panel {
    width: 100%;
    min-width: 600px;
  }

  /* Loading Skeleton */
  .tokens-skeleton {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  .skeleton-toolbar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 1rem;
  }

  .skeleton-search {
    width: 260px;
    height: 36px;
    border-radius: 6px;
  }

  .skeleton-button {
    width: 120px;
    height: 36px;
    border-radius: 6px;
  }

  .skeleton-table {
    display: flex;
    flex-direction: column;
    gap: 0;
    border: 1px solid var(--color-border, #e5e7eb);
    border-radius: 8px;
    overflow: hidden;
  }

  .skeleton-row {
    display: flex;
    gap: 1rem;
    padding: 0.75rem 1rem;
    border-bottom: 1px solid var(--color-border, #e5e7eb);
  }

  .skeleton-row:last-child {
    border-bottom: none;
  }

  .skeleton-header-row {
    background: var(--color-surface-secondary, #f9fafb);
  }

  .skeleton-cell {
    flex: 1;
    height: 16px;
    border-radius: 4px;
  }

  .skeleton-block {
    background: linear-gradient(
      90deg,
      var(--color-skeleton-base, #e5e7eb) 25%,
      var(--color-skeleton-shine, #f3f4f6) 50%,
      var(--color-skeleton-base, #e5e7eb) 75%
    );
    background-size: 200% 100%;
    animation: skeleton-shimmer 1.5s ease-in-out infinite;
  }

  @keyframes skeleton-shimmer {
    0% {
      background-position: 200% 0;
    }
    100% {
      background-position: -200% 0;
    }
  }

  /* Error State */
  .tokens-error {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    text-align: center;
    padding: 3rem 1.5rem;
    gap: 0.75rem;
    min-height: 300px;
  }

  .tokens-error-icon {
    color: var(--color-text-tertiary, #9ca3af);
    margin-bottom: 0.5rem;
  }

  .tokens-error-title {
    font-size: 1.125rem;
    font-weight: 600;
    color: var(--color-text-primary, #111827);
    margin: 0;
  }

  .tokens-error-message {
    font-size: 0.875rem;
    color: var(--color-text-secondary, #6b7280);
    margin: 0 0 0.5rem;
    max-width: 400px;
    line-height: 1.5;
  }

  /* Responsive */
  @media (max-width: 640px) {
    .tokens-page {
      padding: 1rem;
      gap: 1rem;
    }

    .tokens-header {
      flex-direction: column;
    }

    .tokens-header-actions {
      width: 100%;
    }

    .tokens-header-actions :global(button) {
      width: 100%;
    }

    .tokens-tab {
      flex: 1;
      text-align: center;
    }

    .tokens-panel {
      min-width: 0;
    }
  }
</style>