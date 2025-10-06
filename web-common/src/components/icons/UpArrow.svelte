<script lang="ts">
  export let size: string | number = "1em";
  export let color: string = "currentColor";
  export let className: string = "";
</script>

<svg
  width={size}
  height={size}
  viewBox="0 0 24 24"
  fill="none"
  xmlns="http://www.w3.org/2000/svg"
  class={className}
>
  <g transform="translate(5, 5)">
    <path
      d="M6.46399 0.507812C6.79785 0.235541 7.2905 0.255217 7.60168 0.566406L12.1 5.06C12.4316 5.39195 12.4318 5.93029 12.1 6.26215C11.7681 6.59383 11.2298 6.59373 10.8978 6.26215L7.85071 3.21875V12.834C7.85041 13.303 7.4692 13.6834 7.00012 13.6836C6.53102 13.6834 6.15081 13.3031 6.15051 12.834V3.21875L3.1032 6.26215C2.77144 6.59356 2.23392 6.59336 1.90203 6.26215C1.57008 5.93021 1.57008 5.39195 1.90203 5.06L6.39954 0.566406L6.46399 0.507812Z"
      fill={color}
    />
  </g>
</svg>
