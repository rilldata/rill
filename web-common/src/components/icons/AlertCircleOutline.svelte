<script>
  import TopToBottomGradient from "@rilldata/web-common/components/icons/TopToBottomGradient.svelte";

  export let size = "1em";
  export let color = "currentColor";
  export let className = "";

  /**
   * Used to set gradient to icon. Uses `color` as gradient start.
   * Make sure entries are added in app.css before referencing these (Add `--color-` prefix to these labels)
   */
  export let gradientStopColor = "";
  $: hasGradient = !!gradientStopColor;
</script>

<svg
  height={size}
  viewBox="0 0 24 24"
  fill={color}
  xmlns="http://www.w3.org/2000/svg"
  class={className}
>
  <TopToBottomGradient endColor={gradientStopColor} />

  <path
    d="M11 6.75C11 6.47386 11.2239 6.25 11.5 6.25H12.5C12.7761 6.25 13 6.47386 13 6.75V13.25C13 13.5261 12.7761 13.75 12.5 13.75H11.5C11.2239 13.75 11 13.5261 11 13.25V6.75Z"
    fill={hasGradient ? "url(#top-bottom-grad)" : color}
  />
  <path
    d="M12 17.75C12.6904 17.75 13.25 17.1904 13.25 16.5C13.25 15.8096 12.6904 15.25 12 15.25C11.3096 15.25 10.75 15.8096 10.75 16.5C10.75 17.1904 11.3096 17.75 12 17.75Z"
    fill={hasGradient ? "url(#top-bottom-grad)" : color}
  />
  <path
    fill-rule="evenodd"
    clip-rule="evenodd"
    d="M12 22C17.5228 22 22 17.5228 22 12C22 6.47715 17.5228 2 12 2C6.47715 2 2 6.47715 2 12C2 17.5228 6.47715 22 12 22ZM12 20C16.4183 20 20 16.4183 20 12C20 7.58172 16.4183 4 12 4C7.58172 4 4 7.58172 4 12C4 16.4183 7.58172 20 12 20Z"
    fill={hasGradient ? "url(#top-bottom-grad)" : color}
  />
</svg>
