<script>
  // values: números; max: topo da escala (ex.: 100 para %)
  let { values = [], max = 100, color = 'var(--blue)', slots = 24 } = $props();

  const points = $derived.by(() => {
    if (!values.length) return '';
    const vals = values.length === 1 ? [values[0], values[0]] : values;
    const top = Math.max(max, ...vals, 1);
    const n = Math.max(slots, vals.length);
    const step = 120 / (n - 1);
    const offset = (n - vals.length) * step;
    return vals
      .map((v, i) => `${(offset + i * step).toFixed(1)},${(30 - (v / top) * 26).toFixed(1)}`)
      .join(' ');
  });
</script>

<svg viewBox="0 0 120 32" preserveAspectRatio="none" aria-hidden="true">
  <polyline points="0,30 120,30" fill="none" stroke="var(--line)" stroke-width="1" vector-effect="non-scaling-stroke" />
  {#if points}
    <polyline
      {points}
      fill="none"
      stroke={color}
      stroke-width="2"
      vector-effect="non-scaling-stroke"
      stroke-linejoin="round"
      stroke-linecap="round"
    />
  {/if}
</svg>

<style>
  svg { width: 100%; height: 32px; display: block; }
</style>
