<script lang="ts">
  export let size = "1em";
  export let gradient = false;
</script>

<svg
  height={size}
  viewBox="0 0 36 36"
  fill="none"
  xmlns="http://www.w3.org/2000/svg"
>
  <rect
    width="36"
    height="36"
    rx="4"
    fill={gradient ? "url(#paint0_linear_1445_29041)" : "#2563EB"}
  />
  <path
    fill-rule="evenodd"
    clip-rule="evenodd"
    d="M23 14V30L19 28V14H23ZM11 14V24L7 22V14H11Z"
    fill="white"
  />
  <path
    fill-rule="evenodd"
    clip-rule="evenodd"
    d="M25 12L29 14V22H25V12ZM13 6L17 8V22H13V6Z"
    fill="white"
  />
  {#if gradient}
    <defs>
      <linearGradient
        id="paint0_linear_1445_29041"
        x1="18"
        y1="0.5"
        x2="18"
        y2="36.5"
        gradientUnits="userSpaceOnUse"
      >
        <stop stop-color="#558AFF" />
        <stop offset="0.359375" stop-color="#4881FF" />
        <stop offset="1" stop-color="#2563EB" />
      </linearGradient>
      <clipPath id="clip0_1445_29041">
        <rect
          width="24"
          height="24"
          fill="white"
          transform="translate(6 6.5)"
        />
      </clipPath>
    </defs>
  {/if}
</svg>
