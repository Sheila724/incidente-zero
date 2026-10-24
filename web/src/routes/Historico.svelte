<script>
  import { onMount } from 'svelte';
  import TopBar from '../lib/TopBar.svelte';
  import { api } from '../lib/api.js';
  import { sev, duration } from '../lib/format.js';
  import { plantao, atualizarPlantao } from '../lib/plantao.svelte.js';

  let status = $state(null);
  let data = $state(null);
  let erro = $state('');

  onMount(async () => {
    atualizarPlantao();
    try {
      const [st, sc] = await Promise.all([api.status(), api.scenarios()]);
      status = st;
      data = sc;
    } catch (e) {
      erro = e.message;
    }
  });

  const historico = $derived(data?.history ?? []);
  const turnos = $derived(plantao.dados?.anteriores ?? []);
  const semPostmortem = $derived(historico.filter((h) => !h.postmortem).length);

  function dataHora(v) {
    return new Date(v).toLocaleString('pt-BR', { day: '2-digit', month: '2-digit', hour: '2-digit', minute: '2-digit' });
  }
</script>

<TopBar {status} current="historico" />

<main>
  <section class="hero">
    <span class="mono eyebrow">Histórico</span>
    <h1>O que você já atendeu</h1>
    <p class="muted">Incidentes resolvidos, postmortems e turnos de plantão.</p>
  </section>

  {#if erro}
    <p class="erro">{erro}</p>
  {:else if data}
    <section class="panel bloco">
      <div class="head">
        <h2>Incidentes resolvidos</h2>
        {#if semPostmortem}<span class="mono faint">{semPostmortem} sem postmortem</span>{/if}
      </div>
      {#if historico.length}
        <div class="tabela">
          <table>
            <thead>
              <tr>
                <th>Resolvido em</th><th>Incidente</th><th>Severidade</th><th class="r">Tempo</th>
                <th class="r">Dicas</th><th class="r">Pontos</th><th>Postmortem</th>
              </tr>
            </thead>
            <tbody>
              {#each historico as h (h.id)}
                {@const sv = sev(h.severity)}
                <tr>
                  <td class="mono">{dataHora(h.resolvedAt)}</td>
                  <td><a href={`#/lab/${h.id}`}><span class="mono faint">#{h.number}</span> {h.title}</a></td>
                  <td><span class="sevcell"><span class="dot" style:background={sv.color}></span>{sv.label}</span></td>
                  <td class="r mono">{duration(h.seconds)}</td>
                  <td class="r mono">{h.hints}</td>
                  <td class="r mono strong">{h.points}</td>
                  <td>
                    {#if h.postmortem}
                      <a class="pm" href={`#/postmortem/${h.id}`}>Ver postmortem</a>
                    {:else}
                      <a class="pm pend" href={`#/postmortem/${h.id}`}>Escrever</a>
                    {/if}
                  </td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      {:else}
        <p class="muted">Nenhum incidente resolvido ainda. <a href="#/">Escolha um na Central</a>.</p>
      {/if}
    </section>

    <section class="panel bloco">
      <div class="head">
        <h2>Turnos de plantão</h2>
        <a class="mono link" href="#/plantao">Abrir o Plantão</a>
      </div>
      {#if turnos.length}
        <div class="tabela">
          <table>
            <thead>
              <tr>
                <th>Início</th><th>Ritmo</th><th class="r">Incidentes</th><th class="r">Resolvidos</th>
                <th class="r">Escalados</th><th class="r">MTTA</th><th class="r">MTTR</th><th class="r">Pontos</th>
              </tr>
            </thead>
            <tbody>
              {#each turnos as t (t.id)}
                <tr>
                  <td class="mono">{dataHora(t.inicio)}</td>
                  <td>{t.ritmo}</td>
                  <td class="r mono">{t.resumo.incidentes}</td>
                  <td class="r mono">{t.resumo.resolvidos}</td>
                  <td class="r mono">{t.resumo.escalados}</td>
                  <td class="r mono">{t.resumo.mttaSegundos ? duration(t.resumo.mttaSegundos) : '—'}</td>
                  <td class="r mono">{t.resumo.mttrSegundos ? duration(t.resumo.mttrSegundos) : '—'}</td>
                  <td class="r mono strong">{t.resumo.pontos}</td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      {:else}
        <p class="muted">Nenhum turno de plantão ainda. <a href="#/plantao">Assuma o primeiro</a>.</p>
      {/if}
    </section>
  {/if}
</main>

<style>
  main { max-width: 1360px; margin: 0 auto; padding: 36px 32px 56px; display: flex; flex-direction: column; gap: 24px; }
  .eyebrow { font-size: 12px; letter-spacing: .12em; color: var(--accent); text-transform: uppercase; }
  .hero h1 { font-size: 32px; line-height: 1.15; margin: 6px 0; letter-spacing: -0.02em; }
  .hero p { margin: 0; }
  .erro { color: #F9A9AD; }
  .bloco { padding: 22px; border-radius: 14px; display: flex; flex-direction: column; gap: 14px; }
  .bloco p { margin: 0; }
  .bloco p a, .link { color: var(--accent); }
  .head { display: flex; flex-wrap: wrap; align-items: baseline; justify-content: space-between; gap: 8px; }
  h2 { font-size: 18px; margin: 0; }
  .link { font-size: 13px; text-decoration: none; }
  .link:hover { text-decoration: underline; }
  .tabela { overflow-x: auto; }
  table { width: 100%; border-collapse: collapse; min-width: 720px; font-size: 14px; }
  th { text-align: left; font-weight: 500; color: var(--muted); font-size: 12px; padding: 8px 12px; border-bottom: 1px solid var(--line); white-space: nowrap; }
  td { padding: 12px; border-bottom: 1px solid #1A2230; white-space: nowrap; }
  td a { color: var(--text); text-decoration: none; }
  td a:hover { color: var(--accent); }
  .r { text-align: right; }
  .strong { font-weight: 700; }
  .sevcell { display: inline-flex; align-items: center; gap: 8px; }
  .dot { width: 8px; height: 8px; border-radius: 50%; }
  .pm { font-size: 13px; font-weight: 500; color: var(--accent) !important; }
  .pm.pend { color: var(--violet) !important; }
  @media (max-width: 640px) {
    main { padding: 24px 16px 40px; }
    .hero h1 { font-size: 26px; }
  }
</style>
