<script lang="ts">
  export let size = "1em";
</script>

<svg
  height={size}
  width={size}
  viewBox="0 0 24 24"
  fill="none"
  xmlns="http://www.w3.org/2000/svg"
>
  <!-- Globe with connection lines representing HTTPS/Web -->
  <circle cx="12" cy="12" r="9" stroke="#6366F1" stroke-width="1.5" fill="none" />
  <ellipse cx="12" cy="12" rx="4" ry="9" stroke="#6366F1" stroke-width="1.5" fill="none" />
  <path d="M3 12H21" stroke="#6366F1" stroke-width="1.5" />
  <path d="M4.5 7H19.5" stroke="#6366F1" stroke-width="1.5" />
  <path d="M4.5 17H19.5" stroke="#6366F1" stroke-width="1.5" />
</svg>
