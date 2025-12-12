<script>
  export let size = "18px";
  export let primaryColor = "#3524C7";
  export let secondaryColor = "#9CABFF";
</script>

<svg
  xmlns="http://www.w3.org/2000/svg"
  width={size}
  height={size}
  viewBox="0 0 18 18"
  fill="none"
>
  <rect
    x="2"
    y="4"
    width="14"
    height="2"
    rx="1"
    fill={secondaryColor}
    opacity="0.3"
  />
  <rect
    x="2"
    y="8"
    width="12"
    height="2"
    rx="1"
    fill={secondaryColor}
    opacity="0.3"
  />
  <rect
    x="2"
    y="12"
    width="10"
    height="2"
    rx="1"
    fill={secondaryColor}
    opacity="0.3"
  />
  <circle cx="4" cy="5" r="1.5" fill={primaryColor} />
  <circle cx="6" cy="5.5" r="1.5" fill={primaryColor} />
  <circle cx="8" cy="4.5" r="1.5" fill={primaryColor} />
  <circle cx="10" cy="5" r="1.5" fill={primaryColor} />

  <circle cx="4" cy="9" r="1.5" fill={primaryColor} />
  <circle cx="6" cy="9.5" r="1.5" fill={primaryColor} />
  <circle cx="8" cy="8.5" r="1.5" fill={primaryColor} />

  <circle cx="4" cy="13" r="1.5" fill={primaryColor} />
  <circle cx="6" cy="13.5" r="1.5" fill={primaryColor} />
  <circle cx="8" cy="12.5" r="1.5" fill={primaryColor} />
  <path
    fill-rule="evenodd"
    clip-rule="evenodd"
    d="M1.5 15V2.625C1.5 2.41789 1.66789 2.25 1.875 2.25H2.25C2.45711 2.25 2.625 2.41789 2.625 2.625V15H1.5Z"
    fill={primaryColor}
  />
  <path
    fill-rule="evenodd"
    clip-rule="evenodd"
    d="M16.5 14.625C16.5 14.8321 16.3321 15 16.125 15H2.25V13.875H16.125C16.3321 13.875 16.5 14.0429 16.5 14.25V14.625Z"
    fill={primaryColor}
  />
</svg>
