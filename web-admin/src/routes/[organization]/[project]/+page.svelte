<script lang="ts">
  import { page } from "$app/stores";
  import ContentContainer from "@rilldata/web-admin/components/layout/ContentContainer.svelte";
  import DashboardsTable from "@rilldata/web-admin/features/dashboards/listing/DashboardsTable.svelte";
  import InlineChat from "@rilldata/web-common/features/chat/layouts/inline/InlineChat.svelte";
  import DelayedContent from "@rilldata/web-common/features/entity-management/DelayedContent.svelte";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import { createRuntimeServiceGetInstance } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";

  const runtimeClient = useRuntimeClient();
  const { chat } = featureFlags;

  $: ({
    params: { project },
  } = $page);

  $: ({ instanceId } = runtimeClient);

  // Query the instance to get the project display name
  $: instanceQuery = createRuntimeServiceGetInstance(instanceId);
  $: projectDisplayName =
    $instanceQuery.data?.instance?.projectDisplayName || project;
  $: isLoadingDisplayName = $instanceQuery.isLoading;
  $: isErrorDisplayName = $instanceQuery.isError;
</script>

<svelte:head>
  <title>{projectDisplayName} - Rill</title>
</svelte:head>

<ContentContainer maxWidth={900}>
  <div class="flex flex-col gap-y-8 py-12">
    <!-- Welcome Section with Chat Input -->
    <div class="flex flex-col gap-y-6">
      <div class="flex flex-col gap-y-4">
        {#if isLoadingDisplayName}
          <DelayedContent visible={isLoadingDisplayName}>
            <div class="h-11 w-96 animate-pulse rounded bg-gray-200" />
          </DelayedContent>
        {:else if isErrorDisplayName}
          <h1
            class="text-4xl font-semibold text-fg-secondary"
            aria-label="Project title"
          >
            Welcome to <span class="text-accent-primary-action">{project}</span>
          </h1>
        {:else}
          <h1
            class="text-4xl font-semibold text-fg-secondary"
            aria-label="Project title"
          >
            Welcome to <span class="text-accent-primary-action"
              >{projectDisplayName}</span
            >
          </h1>
        {/if}
        <p class="text-lg text-fg-muted">
          {#if $chat}
            Ask questions about your data, or explore your dashboards below
          {:else}
            Explore your dashboards below
          {/if}
        </p>
      </div>

      <!-- Chat Input -->
      {#if $chat}
        <div class="w-full">
          <InlineChat noMargin height="110px" />
        </div>
      {/if}
    </div>

    <!-- Dashboards Section -->
    <div class="flex flex-col gap-y-4">
      <h2 class="text-xl font-semibold text-fg-secondary">Dashboards</h2>
      <DashboardsTable isPreview />
    </div>
  </div>
</ContentContainer>
