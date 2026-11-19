<script lang="ts">
  let {
    size = "12px",
    className = "",
  }: {
    size?: string;
    className?: string;
  } = $props();
</script>

<svg
  width={size}
  height={size}
  viewBox="0 0 12 12"
  fill="none"
  stroke="currentColor"
  stroke-width="1.5"
  xmlns="http://www.w3.org/2000/svg"
  class={className}
>
  <path d="M3 3h6v6M3 9l6-6" />
</svg>
