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
  <!-- Circular background -->
  <circle cx="12" cy="12" r="10" fill={color} />
  <!-- Upward arrow - more slender -->
  <path d="M12 7L9 11H10.75V16H13.25V11H15L12 7Z" fill="white" />
</svg>
