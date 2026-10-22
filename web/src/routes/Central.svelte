<script>
  import { onMount } from 'svelte';
  import TopBar from '../lib/TopBar.svelte';
  import { api } from '../lib/api.js';
  import { sev, LEVELS, duration, today } from '../lib/format.js';

  let status = $state(null);
  let data = $state(null);
  let error = $state('');

  async function load() {
    try {
      const [st, sc] = await Promise.all([api.status(), api.scenarios()]);
      status = st;
      data = sc;
      error = '';
    } catch (e) {
      error = e.message;
    }
  }

  onMount(load);

  const scenarios = $derived(data?.scenarios ?? []);
  // só conta como "em andamento" o lab com tentativa aberta; lab já resolvido não prende o destaque
  const running = $derived(scenarios.find((s) => s.inProgress));
  const next = $derived(scenarios.find((s) => s.id === data?.next));
  const featured = $derived(running ?? next);

  const tracks = $derived(
    LEVELS.map((l) => {
      const list = scenarios.filter((s) => s.level === l.id);
      const done = list.filter((s) => s.resolved).length;
      const total = Math.max(l.planned, list.length);
      return {
        ...l,
        installed: list.length,
        done,
        total,
        // Um segmento por cenário da trilha
        segments: Array.from({ length: total }, (_, i) => i < done)
      };
    })
  );

  // Catálogo agrupado por trilha, na ordem iniciante → intermediário → avançado
  const groups = $derived(
    LEVELS.map((l) => ({ ...l, list: scenarios.filter((s) => s.level === l.id) })).filter(
      (g) => g.list.length
    )
  );
  const TRACK_NAMES = { linux: 'Linux', rede: 'Rede', web: 'Web', banco: 'Banco', docker: 'Docker', zabbix: 'Zabbix', kubernetes: 'Kubernetes' };

  function cardStatus(s) {
    if (s.inProgress) return { label: 'EM ANDAMENTO', cls: 'st-run' };
    if (s.resolved) return { label: 'RESOLVIDO', cls: 'st-ok' };
    if (s.id === data?.next) return { label: 'PRÓXIMO', cls: 'st-next' };
    return { label: 'NOVO', cls: 'st-new' };
  }

  function meta(s) {
    if (s.resolved) return `${duration(s.bestSeconds)} · ${s.bestPoints} pts`;
    return `~${s.minutes} min · ${s.hintCount} dicas`;
  }
</script>

<TopBar {status} current="central" />

<main>
  <section class="hero">
    <div>
      <div class="eyebrow mono">Plantão · {today()}</div>
      <h1>Central de plantão</h1>
      <p class="muted">
        {#if running}
          Há um incidente em andamento. Volte para o lab e termine o que começou.
        {:else}
          Nenhum incidente aberto agora. Escolha o próximo, assuma o plantão e resolva antes que vire manchete.
        {/if}
      </p>
    </div>
    {#if featured}
      <a class="btn btn-primary big" href={`#/lab/${featured.id}`}>
        <svg width="18" height="18" viewBox="0 0 24 24" aria-hidden="true" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 3l9 16H3z"></path><line x1="12" y1="10" x2="12" y2="14"></line><line x1="12" y1="17" x2="12" y2="17.5"></line></svg>
        {running ? 'Voltar ao incidente' : 'Assumir próximo incidente'}
      </a>
    {/if}
  </section>

  {#if error}
    <div class="panel errbox" role="alert">
      <strong>Não foi possível falar com o iz.</strong>
      <span class="muted">{error}</span>
      <button type="button" class="btn" onclick={load}>Tentar de novo</button>
    </div>
  {/if}

  {#if data}
    <section class="stats" aria-label="Seus números">
      <div class="panel stat">
        <span class="muted">Incidentes resolvidos</span>
        <span class="val mono">{data.stats.resolved}<small> / {data.stats.total}</small></span>
        <span class="faint">cenários instalados</span>
      </div>
      <div class="panel stat">
        <span class="muted">MTTR pessoal</span>
        <span class="val mono">{data.stats.resolved ? duration(data.stats.mttrSeconds) : '—'}</span>
        <span class="faint">tempo médio até resolver</span>
      </div>
      <div class="panel stat">
        <span class="muted">Dicas usadas</span>
        <span class="val mono">{data.stats.hintsUsed}</span>
        <span class="faint">nos incidentes resolvidos</span>
      </div>
      <div class="panel stat">
        <span class="muted">Pontuação</span>
        <span class="val mono">{data.stats.points}{#if data.stats.maxPoints}<small> / {data.stats.maxPoints}</small>{/if}</span>
        <span class="faint">melhor resultado de cada incidente</span>
      </div>
    </section>

    <section class="row">
      {#if featured}
        {@const sv = sev(featured.severity)}
        <article class="panel featured">
          <div class="strip" class:crit={featured.severity === 'critica'}>
            <span class="pill" style:background={sv.color} style:color={sv.ink}>{sv.pill}</span>
            <span class="strip-txt">{featured.inProgress ? 'Incidente em andamento' : 'Próximo incidente recomendado'}</span>
            <span class="mono strip-id">#{featured.number} · {featured.track} · {featured.level}</span>
          </div>
          <div class="fbody">
            <h2>{featured.title}</h2>
            <div class="alert mono">[ALERTA] {featured.alertTitle}</div>
            {#if featured.tags?.length}
              <div class="tags">
                {#each featured.tags as t (t)}<span class="chip">{t}</span>{/each}
              </div>
            {/if}
            <div class="factions">
              <a class="btn btn-primary" href={`#/lab/${featured.id}`}>
                {featured.inProgress ? 'Voltar ao plantão' : 'Assumir o plantão'}
              </a>
              <span class="mono faint">~{featured.minutes} min · {featured.hintCount} dicas</span>
            </div>
          </div>
        </article>
      {:else if scenarios.length}
        <article class="panel featured done">
          <div class="fbody">
            <h2>Todos os incidentes instalados foram resolvidos</h2>
            <p class="muted">Refaça algum para melhorar seu tempo, ou contribua com um cenário novo.</p>
          </div>
        </article>
      {/if}

      <article class="panel tracks" id="trilhas">
        <h2>Trilhas</h2>
        {#each tracks as tr (tr.id)}
          <div class="track">
            <div class="track-head">
              <span class="tname">{tr.name}</span>
              {#if !tr.installed}<span class="soon mono">em breve</span>{/if}
              <span class="mono tcount" style:color={tr.done ? tr.color : 'var(--faint)'}>{tr.done} / {tr.total}</span>
            </div>
            <span class="muted small">{tr.desc}</span>
            <div class="segs" aria-hidden="true" style:grid-template-columns={`repeat(${tr.total}, minmax(0, 1fr))`}>
              {#each tr.segments as on, i (i)}
                <span style:background={on ? tr.color : '#1C2532'}></span>
              {/each}
            </div>
          </div>
        {/each}
      </article>
    </section>

    <section class="catalog" id="cenarios">
      {#each groups as g (g.id)}
        <div class="sec-head">
          <h2>{g.title} · {g.desc}</h2>
          <span class="mono faint">{g.list.length} instalados</span>
        </div>
        <div class="grid">
          {#each g.list as s (s.id)}
            {@const st = cardStatus(s)}
            {@const sv = sev(s.severity)}
            <a class="panel card" class:next={st.cls === 'st-next'} class:run={st.cls === 'st-run'} href={`#/lab/${s.id}`}>
              <div class="card-top">
                <span class="mono faint small">#{s.number}</span>
                <span class="dot" style:background={sv.color}></span>
                <span class="muted small">{sv.label}</span>
                {#if g.id !== 'iniciante' && TRACK_NAMES[s.track]}
                  <span class="mono faint small">· {TRACK_NAMES[s.track]}</span>
                {/if}
                <span class="pill {st.cls}">{st.label}</span>
              </div>
              <div class="card-title">{s.title}</div>
              <div class="mono faint small card-meta">{meta(s)}</div>
            </a>
          {/each}
        </div>
      {/each}
    </section>

    <section class="panel history">
      <div class="sec-head"><h2>Histórico de plantões</h2></div>
      {#if data.history?.length}
        <div class="table-wrap">
          <table>
            <thead>
              <tr>
                <th>Incidente</th><th>Severidade</th><th class="r">Tempo</th>
                <th class="r">Dicas</th><th class="r">Pontos</th><th>Status</th><th>Postmortem</th>
              </tr>
            </thead>
            <tbody>
              {#each data.history as h (h.id)}
                {@const sv = sev(h.severity)}
                <tr>
                  <td><a href={`#/lab/${h.id}`}><span class="mono faint">#{h.number}</span> {h.title}</a></td>
                  <td><span class="sevcell"><span class="dot" style:background={sv.color}></span>{sv.label}</span></td>
                  <td class="r mono">{duration(h.seconds)}</td>
                  <td class="r mono">{h.hints}</td>
                  <td class="r mono strong">{h.points}</td>
                  <td><span class="pill st-ok">RESOLVIDO</span></td>
                  <td>
                    {#if h.postmortem}
                      <a class="pm-link" href={`#/postmortem/${h.id}`}>Ver postmortem</a>
                    {:else}
                      <a class="pm-link pend" href={`#/postmortem/${h.id}`}>Escrever</a>
                    {/if}
                  </td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      {:else}
        <p class="muted">Nenhum plantão concluído ainda. O primeiro incidente resolvido aparece aqui.</p>
      {/if}
    </section>

    <footer class="mono faint">
      <span>iz {status?.version ?? ''}</span>
      <span>runtime: {status?.runtime || '—'}</span>
      <span>{data.stats.total} cenários instalados</span>
    </footer>
  {/if}
</main>

<style>
  main { max-width: 1360px; margin: 0 auto; padding: 36px 32px 56px; display: flex; flex-direction: column; gap: 28px; }
  .small { font-size: 12px; }
  .strong { font-weight: 700; }

  .hero { display: flex; flex-wrap: wrap; align-items: flex-end; justify-content: space-between; gap: 20px; }
  .eyebrow { font-size: 12px; letter-spacing: .12em; color: var(--accent); text-transform: uppercase; }
  .hero h1 { font-size: 34px; line-height: 1.15; margin: 6px 0; letter-spacing: -0.02em; }
  .hero p { margin: 0; max-width: 620px; }
  .big { padding: 0 20px; min-height: 48px; }

  .errbox { display: flex; flex-wrap: wrap; gap: 12px; align-items: center; padding: 16px 20px; border-color: var(--red-line); }

  .stats { display: grid; grid-template-columns: repeat(auto-fit, minmax(230px, 1fr)); gap: 16px; }
  .stat { padding: 18px 20px; display: flex; flex-direction: column; gap: 6px; font-size: 13px; }
  .val { font-size: 30px; font-weight: 700; letter-spacing: -0.02em; color: var(--text); line-height: 1.2; }
  .val small { font-size: 14px; color: var(--faint); font-weight: 400; }

  .row { display: flex; flex-wrap: wrap; gap: 20px; align-items: stretch; }
  .featured { flex: 999 1 560px; min-width: 0; overflow: hidden; display: flex; flex-direction: column; border-radius: 14px; }
  .strip {
    display: flex; flex-wrap: wrap; align-items: center; gap: 10px; padding: 14px 22px;
    background: var(--violet-soft); border-bottom: 1px solid var(--violet-line);
  }
  .strip.crit { background: var(--red-soft); border-bottom-color: var(--red-line); }
  .strip-txt { font-size: 13px; color: #CFC2F5; }
  .strip.crit .strip-txt { color: #F9A9AD; }
  .strip-id { margin-left: auto; font-size: 12px; color: #A99BD6; }
  .fbody { padding: 22px; display: flex; flex-direction: column; gap: 16px; flex-grow: 1; }
  .fbody h2 { font-size: 24px; line-height: 1.25; }
  .fbody p { margin: 0; }
  .alert {
    font-size: 13px; color: #C4B5FD; background: #0B0F15; border: 1px solid #222C3A;
    border-radius: 8px; padding: 12px 14px;
  }
  .tags { display: flex; flex-wrap: wrap; gap: 8px; }
  .factions { margin-top: auto; display: flex; flex-wrap: wrap; align-items: center; gap: 12px; }
  .factions .mono { margin-left: auto; font-size: 13px; }

  .tracks { flex: 1 1 340px; min-width: 0; padding: 22px; display: flex; flex-direction: column; gap: 18px; border-radius: 14px; }
  .tracks h2 { font-size: 18px; }
  .track { display: flex; flex-direction: column; gap: 8px; padding-bottom: 16px; border-bottom: 1px solid #1C2532; }
  .track:last-child { border-bottom: none; padding-bottom: 0; }
  .track-head { display: flex; align-items: center; gap: 10px; }
  .tname { font-weight: 600; }
  .soon { font-size: 11px; color: var(--faint); border: 1px solid #263142; border-radius: 4px; padding: 0 6px; }
  .tcount { margin-left: auto; font-size: 13px; }
  .segs { display: grid; gap: 4px; }
  .segs span { height: 6px; border-radius: 2px; }

  .catalog { display: flex; flex-direction: column; gap: 14px; }
  .catalog .grid + .sec-head { margin-top: 18px; }
  .sec-head { display: flex; flex-wrap: wrap; align-items: baseline; justify-content: space-between; gap: 8px; }
  .sec-head h2 { font-size: 18px; }
  .grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(250px, 1fr)); gap: 14px; }
  .card {
    display: flex; flex-direction: column; gap: 12px; padding: 16px 18px; min-height: 128px;
    text-decoration: none; color: var(--text); transition: border-color .15s, transform .15s;
  }
  .card:hover { border-color: #3A4659; color: var(--text); transform: translateY(-1px); }
  .card.next { border-color: var(--accent-line); }
  .card.run { border-color: var(--red-line); }
  .card-top { display: flex; align-items: center; gap: 8px; }
  .card-top .pill { margin-left: auto; }
  .card-title { font-weight: 600; line-height: 1.35; }
  .card-meta { margin-top: auto; }

  .st-ok { color: var(--green); background: var(--green-soft); }
  .st-next { color: var(--accent-ink); background: var(--accent); }
  .st-new { color: var(--muted); background: var(--chip); }
  .st-run { color: #1C0507; background: var(--red); }

  .history { padding: 22px; display: flex; flex-direction: column; gap: 14px; border-radius: 14px; }
  .history p { margin: 0; }
  .table-wrap { overflow-x: auto; }
  table { width: 100%; border-collapse: collapse; min-width: 680px; font-size: 14px; }
  th {
    text-align: left; color: var(--faint); font-size: 12px; text-transform: uppercase; letter-spacing: .06em;
    font-weight: 600; padding: 10px 12px; border-bottom: 1px solid #222C3A;
  }
  td { padding: 14px 12px; border-bottom: 1px solid #1A2230; }
  td a { color: var(--text); text-decoration: none; }
  td a:hover { color: var(--accent); }
  .r { text-align: right; }
  .sevcell { display: inline-flex; align-items: center; gap: 8px; }

  footer { display: flex; flex-wrap: wrap; gap: 8px 24px; font-size: 12px; }

  @media (max-width: 640px) {
    main { padding: 24px 16px 40px; }
    .hero h1 { font-size: 28px; }
  }
  .pm-link { font-size: 13px; font-weight: 500; text-decoration: none; white-space: nowrap; color: var(--accent); }
  .pm-link:hover { text-decoration: underline; }
  .pm-link.pend { color: var(--violet); }
</style>
