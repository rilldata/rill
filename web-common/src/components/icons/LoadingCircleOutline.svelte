<script lang="ts">
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
  fill="none"
  xmlns="http://www.w3.org/2000/svg"
  class={className}
>
  <TopToBottomGradient endColor={gradientStopColor} />

  <path
    d="M12,1A11,11,0,1,0,23,12,11,11,0,0,0,12,1Zm0,19a8,8,0,1,1,8-8A8,8,0,0,1,12,20Z"
    fill={hasGradient ? "url(#top-bottom-grad)" : color}
    opacity=".25"
  />
  <path
    d="M12,4a8,8,0,0,1,7.89,6.7A1.53,1.53,0,0,0,21.38,12h0a1.5,1.5,0,0,0,1.48-1.75,11,11,0,0,0-21.72,0A1.5,1.5,0,0,0,2.62,12h0a1.53,1.53,0,0,0,1.49-1.3A8,8,0,0,1,12,4Z"
    fill={hasGradient ? "url(#top-bottom-grad)" : color}
  >
    <animateTransform
      attributeName="transform"
      dur="0.75s"
      repeatCount="indefinite"
      type="rotate"
      values="0 12 12;360 12 12"
    />
  </path>
</svg>
