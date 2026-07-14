<script lang="ts">
  let {
    size = "16",
    color = "currentColor",
    className = "",
  }: {
    size?: string;
    color?: string;
    className?: string;
  } = $props();
</script>

<svg
  width={size}
  height={size}
  viewBox="0 0 16 16"
  fill="none"
  xmlns="http://www.w3.org/2000/svg"
  class={className}
  aria-label="Filter"
>
  <path
    d="M2.584 2.583v.376l4.583 5.417v3.29l1.666 1.21V8.376l4.584-5.417v-.376H2.584Z"
    stroke={color}
    stroke-width="1.33"
    stroke-linejoin="round"
  />
</svg>
