<script lang="ts">
  import ContentContainer from "@rilldata/web-common/components/layout/ContentContainer.svelte";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import type { LayoutData } from "../$types";
  import LocalInlineChat from "./LocalInlineChat.svelte";
  import DashboardList from "../preview/DashboardList.svelte";

  export let data: LayoutData;

  const { chat } = featureFlags;
</script>

<svelte:head>
  <title>Home - Rill</title>
</svelte:head>

{#if data.initialized}
  <ContentContainer maxWidth={900} showTitle={false}>
    <div class="flex flex-col gap-y-8 py-12">
      <!-- Welcome Section with Chat Input -->
      <div class="flex flex-col gap-y-6">
        <div class="flex flex-col gap-y-4">
          <h1
            class="text-4xl font-semibold text-fg-secondary"
            aria-label="Project title"
          >
            Welcome to <span class="text-accent-primary-action"
              >Rill Cloud Preview</span
            >
          </h1>
          <p class="text-lg text-fg-muted">
            {#if $chat}
              Ask questions about your data, get insights, and explore your
              dashboards with our new chat feature!
            {:else}
              Explore your dashboards below
            {/if}
          </p>
        </div>

        <!-- Chat Input -->
        {#if $chat}
          <div class="w-full">
            <LocalInlineChat noMargin height="110px" />
          </div>
        {/if}
      </div>

      <!-- Dashboards Section -->
      <div class="flex flex-col gap-y-4">
        <h2 class="text-xl font-semibold text-fg-secondary">Dashboards</h2>
        <DashboardList limit={5} showSeeAll />
      </div>
    </div>
  </ContentContainer>
{/if}
