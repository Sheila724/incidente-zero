<script>
  let { size = 30 } = $props();
</script>

<svg width={size} height={size} viewBox="0 0 32 32" aria-hidden="true">
  <circle cx="16" cy="16" r="13" fill="none" stroke="var(--accent)" stroke-width="2.5" />
  <polyline
    points="6,17 11,17 13.5,11 17,22 19.5,14 21.5,17 26,17"
    fill="none"
    stroke="var(--red)"
    stroke-width="2"
    stroke-linejoin="round"
    stroke-linecap="round"
  />
</svg>
