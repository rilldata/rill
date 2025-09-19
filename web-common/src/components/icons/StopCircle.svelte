<script lang="ts">
  export let size: string | number = "1em";
  export let color = "currentColor";
  export let className = "";
</script>

<svg
  width={size}
  height={size}
  viewBox="0 0 24 24"
  fill="none"
  xmlns="http://www.w3.org/2000/svg"
  class={className}
>
  <!-- Outer circle to match rounded icon style in the app -->
  <circle cx="12" cy="12" r="10" fill={color} />
  <!-- Stop square with gentle rounding for a sleeker look -->
  <rect x="8" y="8" width="8" height="8" rx="2" fill="white" />
</svg>
