<script lang="ts">
  import { fly } from "svelte/transition";

  import ProjectCards from "./ProjectCards.svelte";
  import TitleContent from "./TitleContent.svelte";
  import UserTestCta from "./UserTestCTA.svelte";
</script>

<div class="size-fill overflow-x-hidden overflow-y-auto">
  <div
    class="wrapper column p-10 2xl:py-16"
    in:fly|global={{ duration: 1600, delay: 400, y: 8 }}
  >
    <TitleContent />
    <div class="column" in:fly|global={{ duration: 1600, delay: 1200, y: 4 }}>
      <ProjectCards />
      <UserTestCta />
    </div>
  </div>
</div>

<style lang="postcss">
  .wrapper {
    @apply w-screen h-fit min-h-screen;
    @apply bg-no-repeat bg-cover;
    background-image: url("$img/welcome-bg-art.png");
  }

  .column {
    @apply flex flex-col items-center gap-y-6;
  }
</style>
