<script lang="ts">
  export let size = "1em";
  export let color = "currentColor";
  export let className = "";
</script>

<svg
  height={size}
  viewBox="0 0 24 24"
  fill="none"
  xmlns="http://www.w3.org/2000/svg"
  class={className}
>
  <path
    fill-rule="evenodd"
    clip-rule="evenodd"
    d="M18.5 9C18.7761 9 19 9.22386 19 9.5V20.5C19 20.7761 18.7761 21 18.5 21H17.5C17.2239 21 17 20.7761 17 20.5V9.5C17 9.22386 17.2239 9 17.5 9H18.5Z"
    fill={className ? "currentColor" : color}
  />
  <path
    fill-rule="evenodd"
    clip-rule="evenodd"
    d="M12.5 3C12.7761 3 13 3.22386 13 3.5V20.5C13 20.7761 12.7761 21 12.5 21H11.5C11.2239 21 11 20.7761 11 20.5V3.5C11 3.22386 11.2239 3 11.5 3H12.5Z"
    fill={className ? "currentColor" : color}
  />
  <path
    fill-rule="evenodd"
    clip-rule="evenodd"
    d="M6.5 13C6.77614 13 7 13.2239 7 13.5V20.5C7 20.7761 6.77614 21 6.5 21H5.5C5.22386 21 5 20.7761 5 20.5V13.5C5 13.2239 5.22386 13 5.5 13H6.5Z"
    fill={className ? "currentColor" : color}
  />
</svg>
