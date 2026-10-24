<script>
  import Logo from './Logo.svelte';
  import { plantao } from './plantao.svelte.js';

  let { status = null, current = 'central' } = $props();

  const turno = $derived(plantao.dados?.turno ?? null);
  const fimTurno = $derived(
    turno ? new Date(turno.fim).toLocaleTimeString('pt-BR', { hour: '2-digit', minute: '2-digit' }) : ''
  );
  const emIncidente = $derived(['alertando', 'reconhecido'].includes(plantao.dados?.incidente?.status));

  const engine = $derived(
    !status
      ? { text: 'verificando…', color: 'var(--muted)' }
      : status.docker
        ? { text: `docker · ${status.runtime} · pronto`, color: 'var(--green)' }
        : { text: 'docker indisponível', color: 'var(--red)' }
  );
</script>

<header>
  <a href="#/" class="brand">
    <Logo />
    <span class="name">Incidente Zero</span>
    {#if status?.version}<span class="ver mono">{status.version}</span>{/if}
  </a>
  <nav aria-label="Principal">
    <a href="#/" class:active={current === 'central'} aria-current={current === 'central' ? 'page' : undefined}>Central</a>
    <a href="#/plantao" class:active={current === 'plantao'} aria-current={current === 'plantao' ? 'page' : undefined}>Plantão</a>
    <a href="#/historico" class:active={current === 'historico'} aria-current={current === 'historico' ? 'page' : undefined}>Histórico</a>
    <a href="#/ajuda" class:active={current === 'ajuda'} aria-current={current === 'ajuda' ? 'page' : undefined}>Como funciona</a>
  </nav>
  {#if turno}
    <a href="#/plantao" class="turno mono" class:incidente={emIncidente} title="Modo Plantão em andamento">
      <span class="tdot" aria-hidden="true"></span>
      {emIncidente ? 'incidente aberto' : `de plantão até ${fimTurno}`}
    </a>
  {/if}
  <a class="gh" href="https://github.com/Sheila724/incidente-zero" target="_blank" rel="noopener noreferrer" aria-label="Código no GitHub" title="Código no GitHub">
    <svg width="20" height="20" viewBox="0 0 24 24" aria-hidden="true" fill="currentColor"><path d="M12 .5a11.5 11.5 0 0 0-3.64 22.41c.58.1.79-.25.79-.56v-2c-3.2.7-3.88-1.37-3.88-1.37-.53-1.33-1.3-1.69-1.3-1.69-1.05-.72.08-.7.08-.7 1.16.08 1.77 1.19 1.77 1.19 1.03 1.77 2.71 1.26 3.37.96.1-.75.4-1.26.73-1.55-2.55-.29-5.24-1.28-5.24-5.7 0-1.26.45-2.29 1.19-3.1-.12-.29-.52-1.46.11-3.05 0 0 .97-.31 3.17 1.18a11 11 0 0 1 5.77 0c2.2-1.49 3.17-1.18 3.17-1.18.63 1.59.23 2.76.11 3.05.74.81 1.19 1.84 1.19 3.1 0 4.43-2.7 5.4-5.26 5.69.41.36.78 1.06.78 2.14v3.17c0 .31.21.67.8.56A11.5 11.5 0 0 0 12 .5z"></path></svg>
  </a>
  <div class="engine mono" title="Motor dos labs">
    <span class="dot" style:background={engine.color}></span>
    {engine.text}
  </div>
</header>

<style>
  header {
    display: flex; flex-wrap: wrap; align-items: center; gap: 12px 28px;
    padding: 12px 32px; border-bottom: 1px solid #1C2532; background: var(--panel-2);
  }
  .brand { display: flex; align-items: center; gap: 10px; text-decoration: none; color: var(--text); }
  .name { font-weight: 700; font-size: 17px; letter-spacing: -0.01em; }
  .ver { font-size: 11px; color: var(--faint); border: 1px solid #263142; border-radius: 4px; padding: 1px 6px; }
  nav { display: flex; flex-wrap: wrap; gap: 4px; }
  nav a {
    color: var(--muted); text-decoration: none; padding: 8px 12px; border-radius: 6px;
    font-size: 15px; min-height: 40px; display: inline-flex; align-items: center;
  }
  nav a:hover { color: var(--text); }
  nav a.active { color: var(--text); font-weight: 600; background: var(--chip); }
  .turno {
    display: flex; align-items: center; gap: 8px; font-size: 12px; text-decoration: none;
    color: var(--accent); border: 1px solid var(--accent-line); background: var(--accent-soft);
    border-radius: 999px; padding: 6px 12px;
  }
  .turno.incidente { color: #F9A9AD; border-color: var(--red-line); background: var(--red-soft); }
  .tdot { width: 8px; height: 8px; border-radius: 50%; background: currentColor; animation: vivo 2s ease-in-out infinite; }
  @keyframes vivo { 0%, 100% { opacity: 1; } 50% { opacity: .35; } }
  .gh { margin-left: auto; color: var(--muted); display: grid; place-items: center; width: 40px; height: 40px; border-radius: 8px; }
  .gh:hover { color: var(--text); background: var(--chip); }
  .engine {
    display: flex; align-items: center; gap: 8px; font-size: 12px;
    color: var(--muted); border: 1px solid #222C3A; border-radius: 999px; padding: 6px 12px;
  }
  @media (prefers-reduced-motion: reduce) { .tdot { animation: none; } }
  @media (max-width: 640px) {
    header { padding: 12px 16px; gap: 10px 12px; }
    nav { order: 3; width: 100%; }
    nav a { padding: 8px 10px; font-size: 14px; }
  }
</style>
