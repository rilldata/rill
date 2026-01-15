<script lang="ts">
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import type { LayoutData } from "../$types";
  import LocalInlineChat from "./LocalInlineChat.svelte";
  import DashboardPreview from "./DashboardPreview.svelte";

  export let data: LayoutData;

  const { chat } = featureFlags;
</script>

<svelte:head>
  <title>Home - Rill</title>
</svelte:head>

{#if data.initialized}
  <div class="h-full w-full flex flex-col bg-white dark:bg-gray-950 overflow-auto">
    <div class="w-full max-w-[900px] mx-auto px-8">
      <div class="flex flex-col gap-y-8 py-12">
        <!-- Welcome Section with Chat Input -->
        <div class="flex flex-col gap-y-6">
          <div class="flex flex-col gap-y-4">
            <h1
              class="text-4xl font-semibold text-gray-900 dark:text-white"
              aria-label="Project title"
            >
              Welcome to <span class="text-primary-600">Rill Developer Preview</span>
            </h1>
            <p class="text-lg text-gray-600 dark:text-gray-400">
              {#if $chat}
               Test your project locally with the Rill Cloud experience. Ask questions and get instant insights using the chat below, or preview a dashboard!
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
          <h2 class="text-xl font-semibold text-gray-900 dark:text-white">
            Dashboards
          </h2>
          <DashboardPreview />
        </div>
      </div>
    </div>
  </div>
{/if}
