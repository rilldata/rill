<script lang="ts">
  import { page } from "$app/stores";
  import ContentContainer from "@rilldata/web-admin/components/layout/ContentContainer.svelte";
  import DashboardsTable from "@rilldata/web-admin/features/dashboards/listing/DashboardsTable.svelte";
  import InlineChat from "@rilldata/web-common/features/chat/layouts/inline/InlineChat.svelte";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";

  const { chat } = featureFlags;

  $: ({
    params: { project },
  } = $page);
</script>

<svelte:head>
  <title>{project} - Rill</title>
</svelte:head>

<ContentContainer maxWidth={900}>
  <div class="flex flex-col gap-y-8 py-12">
    <!-- Welcome Section with Chat Input -->
    <div class="flex flex-col gap-y-6">
      <div class="flex flex-col gap-y-4">
        <h1 class="text-4xl font-semibold text-gray-900">
          Welcome to <span class="text-primary-600">{project}</span>
        </h1>
        <p class="text-lg text-gray-600">
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
      <h2 class="text-xl font-semibold text-gray-900">Dashboards</h2>
      <DashboardsTable isPreview />
    </div>
  </div>
</ContentContainer>
