<script lang="ts">
  import LoadingSpinner from "./LoadingSpinner.svelte";
  import CtaContentContainer from "./calls-to-action/CTAContentContainer.svelte";
  import CtaLayoutContainer from "./calls-to-action/CTALayoutContainer.svelte";
</script>

<CtaLayoutContainer>
  <CtaContentContainer>
    <div class="mt-10">
      <LoadingSpinner />
    </div>
  </CtaContentContainer>
</CtaLayoutContainer>
