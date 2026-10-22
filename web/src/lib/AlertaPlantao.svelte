<script>
  import { api } from './api.js';
  import { sev, clock } from './format.js';
  import { plantao, alternarMudo, atualizarPlantao } from './plantao.svelte.js';

  let agora = $state(Date.now());
  let enviando = $state(false);
  let erro = $state('');
  let botao = $state(null);

  $effect(() => {
    const t = setInterval(() => (agora = Date.now()), 1000);
    return () => clearInterval(t);
  });

  const inc = $derived(plantao.dados?.incidente?.status === 'alertando' ? plantao.dados.incidente : null);
  const sv = $derived(sev(inc?.severity));
  const tocando = $derived(inc ? Math.max(0, (agora + plantao.diferenca - Date.parse(inc.alertaEm)) / 1000) : 0);
  const prazo = $derived(plantao.dados?.prazoReconhecer ?? 300);

  $effect(() => {
    if (inc && botao) botao.focus();
  });

  async function reconhecer() {
    enviando = true;
    erro = '';
    try {
      const r = await api.reconhecer();
      await atualizarPlantao();
      location.hash = `#/lab/${encodeURIComponent(r.incidente.id)}`;
    } catch (e) {
      erro = e.message;
    } finally {
      enviando = false;
    }
  }
</script>

{#if inc}
  <div class="fundo">
    <div class="alerta" role="alertdialog" aria-modal="true" aria-labelledby="alerta-titulo" aria-describedby="alerta-desc">
      <div class="topo">
        <span class="sirene" aria-hidden="true">
          <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M10.3 3.9 1.8 18a2 2 0 0 0 1.7 3h17a2 2 0 0 0 1.7-3L13.7 3.9a2 2 0 0 0-3.4 0z"></path><line x1="12" y1="9" x2="12" y2="13"></line><line x1="12" y1="17" x2="12.01" y2="17"></line></svg>
        </span>
        <span class="mono rotulo">INCIDENTE NO PLANTÃO</span>
        <span class="pill" style:background={sv.color} style:color={sv.ink}>{sv.pill}</span>
      </div>
      <h2 id="alerta-titulo">{inc.alertTitle}</h2>
      <p id="alerta-desc" class="mono faint">#{inc.number} · alerta há {clock(tocando)}</p>
      {#if inc.escalado}
        <p class="escalado">Escalado: o alerta passou de {Math.round(prazo / 60)} minutos sem reconhecimento. O incidente perde pontos.</p>
      {:else}
        <p class="prazo">Reconheça em até <strong class="mono">{clock(Math.max(0, prazo - tocando))}</strong> para não escalar.</p>
      {/if}
      {#if erro}<p class="erro">{erro}</p>{/if}
      <div class="acoes">
        <button type="button" class="btn btn-primary" bind:this={botao} onclick={reconhecer} disabled={enviando}>
          {enviando ? 'Abrindo…' : 'Reconhecer e abrir o lab'}
        </button>
        <button type="button" class="btn btn-ghost" onclick={alternarMudo} aria-pressed={plantao.mudo}>
          {plantao.mudo ? 'Ligar o som' : 'Silenciar o som'}
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  .fundo { position: fixed; inset: 0; z-index: 100; display: grid; place-items: center; padding: 16px; background: rgba(20, 4, 8, 0.78); }
  .alerta {
    width: min(560px, 100%); background: var(--panel); border: 1px solid var(--red); border-radius: 14px;
    padding: 22px; box-shadow: 0 0 0 4px rgba(242, 84, 91, 0.15), 0 24px 60px rgba(0, 0, 0, 0.6);
    animation: pulso 1.6s ease-in-out infinite;
  }
  .topo { display: flex; align-items: center; gap: 10px; margin-bottom: 12px; }
  .sirene { color: var(--red); display: grid; place-items: center; width: 38px; height: 38px; border-radius: 10px; background: var(--red-soft); border: 1px solid var(--red-line); }
  .rotulo { font-size: 12px; letter-spacing: 0.08em; color: #F9A9AD; font-weight: 700; }
  .pill { margin-left: auto; font-size: 11px; font-weight: 700; padding: 3px 8px; border-radius: 6px; font-family: var(--mono); }
  h2 { margin: 0 0 6px; font-size: 21px; line-height: 1.3; }
  p { margin: 0 0 10px; }
  .faint { color: var(--faint); font-size: 13px; }
  .prazo { color: var(--text-2); }
  .escalado { color: #F9A9AD; background: var(--red-soft); border: 1px solid var(--red-line); padding: 8px 10px; border-radius: 8px; }
  .erro { color: #F9A9AD; }
  .acoes { display: flex; flex-wrap: wrap; gap: 8px; margin-top: 14px; }
  .acoes .btn-primary { flex: 1 1 220px; min-height: 44px; font-size: 16px; }
  @keyframes pulso {
    0%, 100% { box-shadow: 0 0 0 4px rgba(242, 84, 91, 0.15), 0 24px 60px rgba(0, 0, 0, 0.6); }
    50% { box-shadow: 0 0 0 10px rgba(242, 84, 91, 0.28), 0 24px 60px rgba(0, 0, 0, 0.6); }
  }
  @media (prefers-reduced-motion: reduce) { .alerta { animation: none; } }
</style>
