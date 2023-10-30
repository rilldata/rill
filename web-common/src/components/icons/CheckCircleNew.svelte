<script lang="ts">
  /**
   * Slightly different that CheckCircle.svelte. This matches more with our design.
   */
  export let size = "1em";
  export let color = "currentColor";
  export let className = "";
</script>

<svg
  fill="none"
  height={size}
  viewBox="0 0 24 24"
  width={size}
  xmlns="http://www.w3.org/2000/svg"
>
  <path
    class={className}
    d="M17.5216 9.45423C17.6892 9.24518 17.6622 8.94119 17.4604 8.76493L16.5248 7.94757C16.3087 7.75879 15.9783 7.79023 15.8017 8.01641L11.1942 13.9171L8.9537 11.6906C8.7593 11.4974 8.44559 11.4967 8.25035 11.6891L7.35844 12.5678C7.16084 12.7625 7.15948 13.0808 7.35541 13.2771L11.2933 17.2236L17.5216 9.45423Z"
    fill={color}
  />
  <path
    class={className}
    clip-rule="evenodd"
    d="M22 12C22 17.5228 17.5228 22 12 22C6.47715 22 2 17.5228 2 12C2 6.47715 6.47715 2 12 2C17.5228 2 22 6.47715 22 12ZM20 12C20 16.4183 16.4183 20 12 20C7.58172 20 4 16.4183 4 12C4 7.58172 7.58172 4 12 4C16.4183 4 20 7.58172 20 12Z"
    fill={color}
    fill-rule="evenodd"
  />
</svg>
