<script lang="ts">
  import CTAContentContainer from "@rilldata/web-common/components/calls-to-action/CTAContentContainer.svelte";
  import CTALayoutContainer from "@rilldata/web-common/components/calls-to-action/CTALayoutContainer.svelte";
</script>

<CTALayoutContainer>
  <CTAContentContainer>
    <slot />
  </CTAContentContainer>
</CTALayoutContainer>
