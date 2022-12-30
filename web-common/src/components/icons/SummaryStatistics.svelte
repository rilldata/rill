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
  <mask id="path-1-inside-1_3106_19148" fill="white">
    <rect x="10" y="8" width="8" height="8" rx="1" />
  </mask>
  <rect
    x="10"
    y="8"
    width="8"
    height="8"
    rx="1"
    stroke={color}
    stroke-width="3"
    mask="url(#path-1-inside-1_3106_19148)"
  />
  <path
    d="M18 11H21.5C21.7761 11 22 11.2239 22 11.5V12.5C22 12.7761 21.7761 13 21.5 13H18V11Z"
    fill={color}
  />
  <path
    d="M2 11.5C2 11.2239 2.22386 11 2.5 11H10V13H2.5C2.22386 13 2 12.7761 2 12.5V11.5Z"
    fill={color}
  />
  <rect x="2" y="9" width="2" height="6" rx="0.5" fill={color} />
  <rect x="20" y="9" width="2" height="6" rx="0.5" fill={color} />
</svg>
