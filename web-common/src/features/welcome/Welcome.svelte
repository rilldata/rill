<script lang="ts">
  import { fly } from "svelte/transition";
  import BackgroundImage from "./BackgroundImage.svelte";
  import ProjectCards from "./ProjectCards.svelte";
  import TitleContent from "./TitleContent.svelte";
  import UserTestCta from "./UserTestCTA.svelte";
</script>

<BackgroundImage>
  <div in:fly|global={{ duration: 1600, delay: 400, y: 8 }}>
    <div class="flex flex-col items-center gap-y-6 py-10 px-10 2xl:py-16">
      <TitleContent />
      <div
        class="flex flex-col items-center gap-y-6"
        in:fly|global={{ duration: 1600, delay: 1200, y: 4 }}
      >
        <ProjectCards />
        <UserTestCta />
      </div>
    </div>
  </div></BackgroundImage
>
