<script lang="ts">
  export let size = "1em";
  export let color = "currentColor";
</script>

<svg
  height={size}
  viewBox="0 0 24 24"
  fill="none"
  xmlns="http://www.w3.org/2000/svg"
>
  <!-- Top node -->
  <rect x="9" y="2" width="6" height="4" rx="1" fill={color} />
  <!-- Bottom-left node -->
  <rect x="3" y="18" width="6" height="4" rx="1" fill={color} />
  <!-- Bottom-right node -->
  <rect x="15" y="18" width="6" height="4" rx="1" fill={color} />
  <!-- Edge: top to bottom-left -->
  <path
    d="M10.5 6 L10.5 11 Q10.5 13 8.5 14 L6 18"
    stroke={color}
    stroke-width="1.5"
    fill="none"
  />
  <!-- Edge: top to bottom-right -->
  <path
    d="M13.5 6 L13.5 11 Q13.5 13 15.5 14 L18 18"
    stroke={color}
    stroke-width="1.5"
    fill="none"
  />
</svg>
