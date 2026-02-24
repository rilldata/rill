<script lang="ts">
  import ProjectCards from "@rilldata/web-common/features/welcome/ProjectCards.svelte";
  import TitleContent from "@rilldata/web-common/features/welcome/TitleContent.svelte";
  import { fly } from "svelte/transition";
  import OnboardingGenerateSampleData from "@rilldata/web-common/features/welcome/OnboardingGenerateSampleData.svelte";
  import ConnectYourDataSmall from "@rilldata/web-common/features/welcome/ConnectYourDataSmall.svelte";
  import NewSourcesForm from "@rilldata/web-common/features/welcome/new-sources/NewSourcesForm.svelte";

  let selectConnector = false;
  let selectedConnectorName: string | null = null;
</script>

<div class="scroll" in:fly={{ duration: 1600, delay: 400, y: 8 }}>
  <div class="wrapper column p-10 2xl:py-16">
    {#if selectConnector}
      <NewSourcesForm bind:selectedConnectorName />
    {:else}
      <TitleContent />

      <div class="flex flex-row gap-x-12">
        <ConnectYourDataSmall
          startConnectorSelection={(name) => {
            selectedConnectorName = name;
            selectConnector = true;
          }}
        />
        <OnboardingGenerateSampleData />
      </div>

      <div class="column" in:fly={{ duration: 1600, delay: 1200, y: 4 }}>
        <ProjectCards />
      </div>
    {/if}
  </div>
</div>

<style lang="postcss">
  .scroll {
    @apply size-full overflow-x-hidden overflow-y-auto;
  }

  .wrapper {
    @apply w-full h-fit min-h-screen bg-no-repeat bg-cover;
    background-image: url("/img/welcome-bg-art.jpg");
  }

  :global(.dark) .wrapper {
    background-image: url("/img/welcome-bg-art-dark.jpg");
  }

  .column {
    @apply flex flex-col items-center gap-y-6;
  }
</style>
