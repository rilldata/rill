<script lang="ts">
  import LoadingCircleOutline from "../icons/LoadingCircleOutline.svelte";

  export let disabled = false;
  export let isLoading = false;
  export let redirect = false;
  export let href = "/";
  export let imageUrl = "";
</script>

<a
  href={href + (redirect ? "?redirect=true" : "")}
  class:gradient={!imageUrl}
  on:click
  on:keydown={(e) => e.key === "Enter" && e.currentTarget.click()}
  aria-disabled={disabled}
  class:loading={isLoading}
  style:background-image={imageUrl ? `url('${imageUrl}')` : ""}
>
  {#if isLoading}
    <div
      class="absolute z-10 inset-0 flex items-center justify-center backdrop-blur-sm"
    >
      <LoadingCircleOutline size="48px" color="var(--color-primary-600)" />
    </div>
  {/if}
  <slot />
</a>

<style lang="postcss">
  a {
    @apply bg-no-repeat bg-center border bg-cover;
    @apply relative select-none;
    @apply size-60 rounded-md;
    @apply flex flex-col items-center justify-center gap-y-2;
    @apply cursor-pointer overflow-hidden;

    box-shadow:
      0px 2px 3px rgba(15, 23, 42, 0.06),
      0px 1px 3px rgba(15, 23, 42, 0.08);
  }

  .gradient {
    @apply bg-gradient-to-b from-[#FFFFFF] to-[#F8FAFC];
  }

  :global(.dark) .gradient {
    @apply bg-gray-300;
    background-image: linear-gradient(#6b6b6b33 34%, #00000033);
  }

  a[aria-disabled="true"] {
    cursor: not-allowed;
    pointer-events: none;
  }

  a[aria-disabled="true"]:not(.loading) {
    opacity: 0.4;
  }

  a:hover {
    @apply shadow-lg;
  }
</style>
