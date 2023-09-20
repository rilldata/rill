<script lang="ts">
  import { fly } from "svelte/transition";
  import BackgroundImage from "./BackgroundImage.svelte";
  import ProjectCards from "./ProjectCards.svelte";
  import TitleContent from "./TitleContent.svelte";
  import UserTestCta from "./UserTestCTA.svelte";
</script>

<BackgroundImage>
  <div in:fly={{ duration: 1600, delay: 400, y: 8 }}>
    <div
      class="flex flex-col items-center gap-y-8 py-16 2xl:py-32 px-10 2xl:px-20"
    >
      <TitleContent />
      <div
        class="flex flex-col items-center"
        in:fly={{ duration: 1600, delay: 1200, y: 4 }}
      >
        <ProjectCards />
        <UserTestCta />
      </div>
    </div>
  </div></BackgroundImage
>
