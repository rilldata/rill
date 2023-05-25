<script lang="ts">
  import { fly } from "svelte/transition";
  import BackgroundImage from "./BackgroundImage.svelte";
  import BodyContainer from "./BodyContainer.svelte";
  import ProjectCards from "./ProjectCards.svelte";
  import TitleContent from "./TitleContent.svelte";
</script>

<BackgroundImage>
  <div in:fly={{ duration: 1600, delay: 400, y: 8 }}>
    <BodyContainer>
      <TitleContent />
      <ProjectCards />
    </BodyContainer>
  </div>
</BackgroundImage>
