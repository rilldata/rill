<script lang="ts">
  import { fly } from "svelte/transition";
  export let size = "1em";
  export let iconOnly = false;
</script>

<svg
  height={size}
  viewBox={iconOnly ? "0 0 560 600" : "0 0 1550 600"}
  version="1.1"
  xmlns="http://www.w3.org/2000/svg"
  xmlns:xlink="http://www.w3.org/1999/xlink"
  ><title>Logo</title><g
    id="Social-Assets"
    stroke="none"
    stroke-width="1"
    fill="none"
    fill-rule="evenodd"
    ><g id="Branding-/-Logos-/-Full-Logo-/-On-Light"
      ><path
        transition:fly|local={{ duration: 400, y: 420 }}
        d="M400,200 L400,600 L300,550 L300,200 L400,200 Z M100,200 L100,450 L0,400 L0,200 L100,200 Z"
        id="Shape"
        fill="#1E78FA"
      /><path
        transition:fly|local={{ duration: 400, y: -410 }}
        d="M450,150 L550,200 L550,400 L450,400 L450,150 Z M150,0 L250,50 L250,400 L150,400 L150,0 Z"
        id="Shape"
        fill="#00C8D7"
      /><path
        transition:fly|local={{ duration: 400, y: 420 }}
        d="M878,50 C972.992977,50 1050,127.007023 1050,222 L1050,228 C1050,290.399114 1016.772,345.037413 967.044921,375.18598 C968.429897,377.636958 969.751702,380.140737 971.005802,382.691474 L971.751171,384.22975 L1050,550 L940,550 L881.510251,427.317757 L881.510251,427.317757 C873.663379,410.883741 857.206798,400.328419 839.047068,400.007529 L838.194609,400 L800,400 L800,500 L700,550 L700,50 L878,50 Z M1200,300 L1200,550 L1100,550 L1100,350 L1200,300 Z M1250,50 L1350,100 L1350,550 L1250,550 L1250,50 Z M1400,125 L1500,175 L1500,550 L1400,550 L1400,125 Z M875,150 L800,150 L800,300 L875,300 C916.421356,300 950,266.421356 950,225 C950,183.578644 916.421356,150 875,150 Z M1150,150 C1177.61424,150 1200,172.385763 1200,200 C1200,227.614237 1177.61424,250 1150,250 C1122.38576,250 1100,227.614237 1100,200 C1100,172.385763 1122.38576,150 1150,150 Z"
        id="Shape"
        fill="#0F284B"
      /></g
    ></g
  ></svg
>
