<script lang="ts">
  import Cancel from "@rilldata/web-common/components/icons/Cancel.svelte";
  import Check from "@rilldata/web-common/components/icons/Check.svelte";
  import Spacer from "@rilldata/web-common/components/icons/Spacer.svelte";
  export let selected = false;
  export let excluded = false;
</script>

<div style:width="22px" class="grid place-items-center">
  {#if selected && !excluded}
    <Check size="20px" />
  {:else if selected && excluded}
    <Cancel size="20px" />
  {:else}
    <Spacer />
  {/if}
</div>
