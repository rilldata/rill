<script lang="ts">
  import { IconSpaceFixer } from "@rilldata/web-common/components/button";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";

  export let active = false;
  export let disabled = false;
</script>

<button
  {disabled}
  class:bg-gray-200={active}
  class="px-3 py-2 rounded grid gap-x-2 hover:bg-gray-200 hover:dark:bg-gray-600 items-center"
  style:grid-template-columns="{$$slots["icon"] ? "max-content" : ""}
  max-content
  {$$slots["context"] ? "max-content" : ""}
  {disabled ? "" : "max-content"}"
  on:click
>
  {#if $$slots["icon"]}
    <span class="ui-copy-icon"><slot name="icon" /></span>
  {/if}
  <span class="font-bold" style:transform="translateY(1px)">
    <slot />
  </span>
  {#if $$slots["context"]}
    <span style:transform="translateY(1px)">
      <slot name="context" />
    </span>
  {/if}

  {#if !disabled}
    <IconSpaceFixer pullRight>
      <div class="transition-transform" class:-rotate-180={active}>
        <CaretDownIcon size="14px" />
      </div>
    </IconSpaceFixer>
  {/if}
</button>
