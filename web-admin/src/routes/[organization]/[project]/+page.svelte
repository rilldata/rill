<script lang="ts">
  import { page } from "$app/stores";
  import { createAdminServiceGetCurrentUser } from "@rilldata/web-admin/client";
  import ContentContainer from "@rilldata/web-admin/components/layout/ContentContainer.svelte";
  import DashboardsPreview from "@rilldata/web-admin/features/dashboards/listing/DashboardsPreview.svelte";
  import InlineChat from "@rilldata/web-common/features/chat/layouts/inline/InlineChat.svelte";

  $: ({
    params: { organization, project },
  } = $page);

  $: currentUser = createAdminServiceGetCurrentUser();
  $: firstName = $currentUser.data?.user?.displayName
    ? $currentUser.data.user.displayName.split(" ")[0]
    : null;
</script>

<svelte:head>
  <title>{project} - Rill</title>
</svelte:head>

<ContentContainer maxWidth={900}>
  <div class="flex flex-col gap-y-16 py-12">
    <!-- Welcome Section with Chat Input -->
    <div class="flex flex-col gap-y-8 pt-8">
      <div class="text-center flex flex-col gap-y-4">
        <h1 class="text-4xl font-semibold text-gray-900">
          {#if firstName}
            Welcome back, {firstName}
          {:else}
            Welcome
          {/if}
        </h1>
        <p class="text-lg text-gray-600">
          Ask questions about your data to get started, or explore your
          dashboards below
        </p>
      </div>

      <!-- Chat Input -->
      <div class="max-w-2xl mx-auto w-full px-4">
        <InlineChat />
      </div>
    </div>

    <!-- Dashboards Section -->
    <DashboardsPreview {organization} {project} />
  </div>
</ContentContainer>
