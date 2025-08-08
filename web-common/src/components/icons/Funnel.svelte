<script lang="ts">
  export let size: number | string = 16;
  export let primaryColor = "#3524C7";
  export let secondaryColor = "#9CABFF";
</script>

<svg
  xmlns="http://www.w3.org/2000/svg"
  width={size}
  height={size}
  viewBox="0 0 24 24"
  fill="none"
>
  <path
    d="M20.5 5C20.7761 5 21 5.22386 21 5.5V7C21 7.27614 20.7761 7.5 20.5 7.5H3.5C3.22386 7.5 3 7.27614 3 7V5.5C3 5.22386 3.22386 5 3.5 5H20.5Z"
    fill={primaryColor}
  />
  <path
    d="M18.5 9C18.7761 9 19 9.22386 19 9.5V11C19 11.2761 18.7761 11.5 18.5 11.5H5.5C5.22386 11.5 5 11.2761 5 11V9.5C5 9.22386 5.22386 9 5.5 9H18.5Z"
    fill={secondaryColor}
  />
  <path
    d="M15.5 13C15.7761 13 16 13.2239 16 13.5V15C16 15.2761 15.7761 15.5 15.5 15.5H8.5C8.22386 15.5 8 15.2761 8 15V13.5C8 13.2239 8.22386 13 8.5 13H15.5Z"
    fill={primaryColor}
  />
  <path
    d="M13.5 17C13.7761 17 14 17.2239 14 17.5V19C14 19.2761 13.7761 19.5 13.5 19.5H10.5C10.2239 19.5 10 19.2761 10 19V17.5C10 17.2239 10.2239 17 10.5 17H13.5Z"
    fill={secondaryColor}
  />
</svg>
