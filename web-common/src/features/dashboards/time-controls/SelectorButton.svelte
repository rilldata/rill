<script lang="ts">
  import { IconSpaceFixer } from "@rilldata/web-common/components/button";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import type { Builder } from "bits-ui";
  import { builderActions, getAttrs } from "bits-ui";

  export let active = false;
  export let disabled = false;
  export let label: string | undefined = undefined;
  export let builders: Builder[] = [];
</script>

<button
  {...getAttrs(builders)}
  use:builderActions={{ builders }}
  {disabled}
  aria-label={label}
  class:bg-gray-200={active}
  class="px-3 py-2 rounded grid gap-x-2 {!disabled
    ? 'hover:bg-gray-200'
    : ''}  items-center"
  style:grid-template-columns="{$$slots['icon'] ? 'max-content' : ''}
  max-content
  {$$slots['context'] ? 'max-content' : ''}
  {disabled ? '' : 'max-content'}"
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
