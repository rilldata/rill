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
    d="M16.5476 14.99C16.8754 14.9694 17.1271 15.3143 16.9311 15.5752C15.8263 17.0466 14.0568 18 12.0622 18C8.71415 18 6 15.3137 6 12C6 8.68629 8.71415 6 12.0622 6C12.2241 6 12.3846 6.00628 12.5433 6.01862C12.8708 6.04406 12.9676 6.45824 12.736 6.68886C11.8512 7.57018 11.3044 8.7844 11.3044 10.125C11.3044 12.8174 13.5097 15 16.23 15C16.3367 15 16.4426 14.9966 16.5476 14.99Z"
    fill={hasGradient ? "url(#top-bottom-grad)" : color}
  />
  <path
    fill-rule="evenodd"
    clip-rule="evenodd"
    d="M22 12C22 17.5228 17.5228 22 12 22C6.47715 22 2 17.5228 2 12C2 6.47715 6.47715 2 12 2C17.5228 2 22 6.47715 22 12ZM20 12C20 16.4183 16.4183 20 12 20C7.58172 20 4 16.4183 4 12C4 7.58172 7.58172 4 12 4C16.4183 4 20 7.58172 20 12Z"
    fill={hasGradient ? "url(#top-bottom-grad)" : color}
  />
</svg>
