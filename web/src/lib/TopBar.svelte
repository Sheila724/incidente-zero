<script>
  import Logo from './Logo.svelte';
  import { plantao } from './plantao.svelte.js';

  let { status = null, current = 'central' } = $props();

  function scrollTo(id) {
    if (location.hash && location.hash !== '#/') {
      location.hash = '#/';
      setTimeout(() => document.getElementById(id)?.scrollIntoView({ behavior: 'smooth' }), 60);
    } else {
      document.getElementById(id)?.scrollIntoView({ behavior: 'smooth' });
    }
  }

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
    <a href="#/" class:active={current === 'central'}>Central</a>
    <button type="button" onclick={() => scrollTo('trilhas')}>Trilhas</button>
    <button type="button" onclick={() => scrollTo('cenarios')}>Cenários</button>
    <a href="#/plantao" class:active={current === 'plantao'}>Plantão</a>
  </nav>
  {#if turno}
    <a href="#/plantao" class="turno mono" class:incidente={emIncidente} title="Modo Plantão em andamento">
      <span class="tdot" aria-hidden="true"></span>
      {emIncidente ? 'incidente aberto' : `de plantão até ${fimTurno}`}
    </a>
  {/if}
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
  nav a, nav button {
    color: var(--muted); text-decoration: none; padding: 8px 12px; border-radius: 6px;
    background: none; border: none; cursor: pointer; font-size: 15px; min-height: 40px;
  }
  nav a:hover, nav button:hover { color: var(--text); }
  nav a.active { color: var(--text); font-weight: 600; background: var(--chip); }
  .turno {
    display: flex; align-items: center; gap: 8px; font-size: 12px; text-decoration: none;
    color: var(--accent); border: 1px solid var(--accent-line); background: var(--accent-soft);
    border-radius: 999px; padding: 6px 12px;
  }
  .turno.incidente { color: #F9A9AD; border-color: var(--red-line); background: var(--red-soft); }
  .tdot { width: 8px; height: 8px; border-radius: 50%; background: currentColor; animation: vivo 2s ease-in-out infinite; }
  @keyframes vivo { 0%, 100% { opacity: 1; } 50% { opacity: .35; } }
  @media (prefers-reduced-motion: reduce) { .tdot { animation: none; } }
  .engine {
    margin-left: auto; display: flex; align-items: center; gap: 8px; font-size: 12px;
    color: var(--muted); border: 1px solid #222C3A; border-radius: 999px; padding: 6px 12px;
  }
  @media (max-width: 640px) {
    header { padding: 12px 16px; }
    .turno {
    display: flex; align-items: center; gap: 8px; font-size: 12px; text-decoration: none;
    color: var(--accent); border: 1px solid var(--accent-line); background: var(--accent-soft);
    border-radius: 999px; padding: 6px 12px;
  }
  .turno.incidente { color: #F9A9AD; border-color: var(--red-line); background: var(--red-soft); }
  .tdot { width: 8px; height: 8px; border-radius: 50%; background: currentColor; animation: vivo 2s ease-in-out infinite; }
  @keyframes vivo { 0%, 100% { opacity: 1; } 50% { opacity: .35; } }
  @media (prefers-reduced-motion: reduce) { .tdot { animation: none; } }
  .engine { margin-left: 0; }
  }
</style>
