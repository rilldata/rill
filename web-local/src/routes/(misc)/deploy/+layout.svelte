<script lang="ts">
  import CTAContentContainer from "@rilldata/web-common/components/calls-to-action/CTAContentContainer.svelte";
  import CTALayoutContainer from "@rilldata/web-common/components/calls-to-action/CTALayoutContainer.svelte";
</script>

<!-- TODO: mocks have the background as white without any header.
           Need a confirmation on this since it is different from current deploy pages. -->
<div class="bg-white">
  <CTALayoutContainer>
    <CTAContentContainer>
      <slot />
    </CTAContentContainer>
  </CTALayoutContainer>
</div>
