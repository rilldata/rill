<script lang="ts">
  export let size: string | number = "1em";
  export let color: string = "currentColor";
  export let className: string = "";
</script>

<svg
  width={size}
  height={size}
  viewBox="0 0 14 14"
  fill="none"
  xmlns="http://www.w3.org/2000/svg"
  class={className}
>
  <path
    fill-rule="evenodd"
    clip-rule="evenodd"
    d="M6.41667 11.8125L6.41667 7.875V7.58333H6.125H2.1875C1.94588 7.58333 1.75 7.45275 1.75 7.29167V6.70833C1.75 6.54725 1.94588 6.41667 2.1875 6.41667H6.125H6.41667V6.125L6.41667 2.1875C6.41667 1.94588 6.54725 1.75 6.70833 1.75H7.29167C7.45275 1.75 7.58333 1.94588 7.58333 2.1875V6.125V6.41667H7.875H11.8125C12.0541 6.41667 12.25 6.54725 12.25 6.70833V7.29167C12.25 7.45275 12.0541 7.58333 11.8125 7.58333H7.875H7.58333V7.875V11.8125C7.58333 12.0541 7.45275 12.25 7.29167 12.25H6.70833C6.54725 12.25 6.41667 12.0541 6.41667 11.8125Z"
    fill={color}
  />
</svg>
