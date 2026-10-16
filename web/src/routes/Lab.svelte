<script>
  import { onMount } from 'svelte';
  import TopBar from '../lib/TopBar.svelte';
  import Terminal from '../lib/Terminal.svelte';
  import Markdown from '../lib/Markdown.svelte';
  import Sparkline from '../lib/Sparkline.svelte';
  import { api } from '../lib/api.js';
  import { sev, clock, duration } from '../lib/format.js';

  let { id } = $props();

  let status = $state(null);
  let detail = $state(null);
  let loadError = $state('');

  let phase = $state('loading'); // loading | briefing | starting | lab
  let startError = $state(null);
  let startClock = $state(0);

  let tab = $state('guiado');
  let stepState = $state([]);
  let metrics = $state(null);
  let series = $state({ http: [], cpu: [], mem: [], disk: [] });

  let checking = $state(false);
  let check = $state(null);
  let result = $state(null);

  let hints = $state([]);
  let hintBusy = $state(false);

  let solution = $state(null);
  let confirmSolution = $state(false);
  let confirmStop = $state(false);
  let stopping = $state(false);

  let now = $state(Date.now());

  const sc = $derived(detail?.scenario);
  const sv = $derived(sev(sc?.severity));
  const resolved = $derived(!!result || !!check?.passed);
  const startedAt = $derived(detail?.startedAt ?? null);
  const elapsed = $derived(
    result ? result.seconds : startedAt ? Math.floor((now - startedAt) / 1000) : null
  );
  const hintsRevealed = $derived(hints.filter((h) => h.text).length);

  const steps = $derived(
    (detail?.guided ?? []).map((g, i, all) => {
      const st = stepState[i];
      // passo de investigação (sem verify) conta como feito quando um passo seguinte já passou
      const laterDone = !g.hasVerify && all.some((x, j) => j > i && x.hasVerify && stepState[j] === true);
      return { ...g, done: st === true || laterDone || (resolved && !g.optional) };
    })
  );
  const currentStep = $derived(steps.findIndex((s) => s.hasVerify && !s.done && !s.optional));

  async function load() {
    try {
      const [st, d] = await Promise.all([api.status(), api.scenario(id)]);
      status = st;
      detail = d;
      hints = d.hints;
      phase = d.scenario.running ? 'lab' : 'briefing';
      if (!d.guided?.length) tab = 'briefing';
      if (d.solutionViewed) loadSolution(true);
    } catch (e) {
      loadError = e.message;
    }
  }

  onMount(() => {
    load();
    const t = setInterval(() => (now = Date.now()), 1000);
    const onKey = (e) => {
      if (phase === 'lab' && e.ctrlKey && e.key === 'Enter') {
        e.preventDefault();
        verify();
      }
    };
    window.addEventListener('keydown', onKey, true);
    return () => {
      clearInterval(t);
      window.removeEventListener('keydown', onKey, true);
    };
  });

  // Monitoramento e passos guiados: atualizam sozinhos enquanto o lab roda.
  $effect(() => {
    if (phase !== 'lab') return;
    let alive = true;
    const pollMetrics = async () => {
      try {
        const m = await api.metrics(id);
        if (!alive) return;
        metrics = m;
        const push = (arr, v) => [...arr, v].slice(-24);
        const ok = /^[23]/.test(m.httpCode);
        series = {
          http: push(series.http, ok ? 100 : 4),
          cpu: push(series.cpu, m.cpuPercent),
          mem: push(series.mem, m.memPercent),
          disk: push(series.disk, m.diskPercent)
        };
      } catch {
        /* lab reiniciando */
      }
    };
    const pollSteps = async () => {
      if (!detail?.guided?.some((g) => g.hasVerify)) return;
      try {
        const r = await api.steps(id);
        if (alive) stepState = r.steps;
      } catch {
        /* ignora */
      }
    };
    pollMetrics();
    pollSteps();
    const a = setInterval(pollMetrics, 5000);
    const b = setInterval(pollSteps, 6000);
    return () => {
      alive = false;
      clearInterval(a);
      clearInterval(b);
    };
  });

  async function start() {
    phase = 'starting';
    startError = null;
    const t0 = Date.now();
    const tick = setInterval(() => (startClock = Math.floor((Date.now() - t0) / 1000)), 500);
    try {
      await api.start(id);
      detail = await api.scenario(id);
      hints = detail.hints;
      check = null;
      result = null;
      solution = null;
      series = { http: [], cpu: [], mem: [], disk: [] };
      tab = detail.guided?.length ? 'guiado' : 'briefing';
      phase = 'lab';
    } catch (e) {
      startError = { message: e.message, output: e.body?.output };
      phase = 'briefing';
    } finally {
      clearInterval(tick);
    }
  }

  async function verify() {
    if (checking || phase !== 'lab') return;
    checking = true;
    try {
      const r = await api.check(id);
      check = r;
      if (r.passed && r.result) result = r.result;
      if (r.passed && !r.result && !result) result = { seconds: elapsed ?? 0, hints: hintsRevealed, points: null };
    } catch (e) {
      check = { passed: false, items: [{ ok: false, text: e.message }], at: Date.now() };
    } finally {
      checking = false;
    }
  }

  async function reveal(n) {
    hintBusy = true;
    try {
      const h = await api.hint(id, n);
      hints = hints.map((x) => (x.n === n ? { ...x, text: h.text } : x));
    } finally {
      hintBusy = false;
    }
  }

  async function loadSolution(force = false) {
    if (!force && !resolved && !confirmSolution) {
      confirmSolution = true;
      return;
    }
    confirmSolution = false;
    solution = await api.solution(id);
    tab = 'solucao';
  }

  async function stop() {
    if (!confirmStop) {
      confirmStop = true;
      return;
    }
    stopping = true;
    try {
      await api.stop(id);
    } finally {
      location.hash = '#/';
    }
  }

  const tabs = $derived([
    ['briefing', 'Briefing'],
    ...(detail?.guided?.length ? [['guiado', 'Guiado']] : []),
    ['conceitos', 'Conceitos'],
    ['dicas', `Dicas ${hintsRevealed}/${hints.length}`],
    ...(solution ? [['solucao', 'Solução']] : [])
  ]);

  const monitors = $derived.by(() => {
    if (!metrics) return [];
    const ok = /^[23]/.test(metrics.httpCode);
    const diskColor = metrics.diskPercent >= 90 ? 'var(--red)' : metrics.diskPercent >= 80 ? 'var(--violet)' : 'var(--blue)';
    return [
      { label: 'HTTP /', value: metrics.httpCode, color: ok ? 'var(--green)' : 'var(--red)', series: series.http, max: 100,
        sub: ok ? 'respondendo' : metrics.httpCode === '000' ? 'sem resposta' : 'erro no servidor' },
      { label: 'Porta 80', value: metrics.port80 ? 'ABERTA' : 'FECHADA', color: metrics.port80 ? 'var(--green)' : 'var(--red)',
        series: series.http, max: 100, sub: metrics.port80 ? 'há processo escutando' : 'nada escutando' },
      { label: 'CPU', value: `${metrics.cpuPercent.toFixed(0)}%`, color: metrics.cpuPercent >= 85 ? 'var(--red)' : 'var(--blue)',
        series: series.cpu, max: 100, sub: 'do limite do lab' },
      { label: 'Memória', value: `${metrics.memPercent.toFixed(0)}%`, color: metrics.memPercent >= 90 ? 'var(--red)' : 'var(--blue)',
        series: series.mem, max: 100, sub: metrics.memUsage || '' },
      { label: `Disco ${metrics.diskPath || '/'}`, value: `${metrics.diskPercent}%`, color: diskColor,
        series: series.disk, max: 100, sub: metrics.diskPercent >= 90 ? 'quase cheio' : 'uso da partição' }
    ];
  });
</script>

<TopBar {status} current="lab" />

{#if loadError}
  <main class="center">
    <div class="panel box" role="alert">
      <h1>Não foi possível abrir este incidente</h1>
      <p class="muted">{loadError}</p>
      <a class="btn" href="#/">Voltar à central</a>
    </div>
  </main>
{:else if phase === 'loading' || !detail}
  <main class="center"><span class="spinner" aria-label="Carregando"></span></main>
{:else if phase === 'briefing' || phase === 'starting'}
  <!-- ANTES DE ASSUMIR O PLANTÃO -->
  <main class="pre">
    <a class="back" href="#/">
      <svg width="16" height="16" viewBox="0 0 24 24" aria-hidden="true" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="15 18 9 12 15 6"></polyline></svg>
      Central
    </a>
    <div class="pre-grid">
      <article class="panel pre-main">
        <div class="pre-head">
          <span class="pill" style:background={sv.color} style:color={sv.ink}>{sv.pill}</span>
          <span class="mono faint">#{sc.number} · {sc.track} · {sc.level}</span>
        </div>
        <h1>{sc.title}</h1>
        <Markdown source={detail.briefing} />
        {#if detail.objectives?.length}
          <h2 class="obj-title">O que você vai aprender</h2>
          <ul class="obj">
            {#each detail.objectives as o (o)}<li>{o}</li>{/each}
          </ul>
        {/if}
      </article>

      <aside class="panel pre-side">
        <dl>
          <div><dt>Servidor</dt><dd class="mono">{detail.hostname || '—'}</dd></div>
          <div><dt>Tempo estimado</dt><dd class="mono">~{sc.minutes} min</dd></div>
          <div><dt>Dicas disponíveis</dt><dd class="mono">{sc.hintCount}</dd></div>
          <div><dt>Melhor resultado</dt><dd class="mono">{sc.resolved ? `${duration(sc.bestSeconds)} · ${sc.bestPoints} pts` : '—'}</dd></div>
        </dl>

        {#if phase === 'starting'}
          <div class="starting" role="status">
            <span class="spinner"></span>
            <div>
              <strong>Preparando o servidor do incidente…</strong>
              <span class="muted small">Subindo o lab, configurando o ambiente e injetando a falha. {startClock}s</span>
            </div>
          </div>
        {:else}
          <button type="button" class="btn btn-primary wide" onclick={start}>
            {sc.resolved ? 'Refazer este plantão' : 'Assumir o plantão'}
          </button>
          <p class="faint small">O cronômetro começa quando o lab fica pronto. Outros labs em execução são encerrados.</p>
        {/if}

        {#if startError}
          <div class="start-err" role="alert">
            <strong>Não deu para subir o lab.</strong>
            <span>{startError.message}</span>
            {#if startError.output}<pre>{startError.output}</pre>{/if}
          </div>
        {/if}
      </aside>
    </div>
  </main>
{:else}
  <!-- LAB EM ANDAMENTO -->
  <div class="incident" class:ok={resolved}>
    <a class="back" href="#/">
      <svg width="16" height="16" viewBox="0 0 24 24" aria-hidden="true" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="15 18 9 12 15 6"></polyline></svg>
      Central
    </a>
    {#if resolved}
      <span class="pill" style="background: var(--green); color: #06140D">RESOLVIDO</span>
    {:else}
      <span class="pill pulse" style:background={sv.color} style:color={sv.ink}>{sv.pill}</span>
    {/if}
    <div class="inc-title">
      <h1>{resolved ? `${sc.title} — resolvido` : sc.alertTitle}</h1>
      <span class="mono muted small">#{sc.number} · host {detail.hostname} · {sc.title}</span>
    </div>
    <div class="inc-actions">
      {#if elapsed !== null}
        <div class="timer">
          <span class="mono" class:green={resolved}>{clock(elapsed)}</span>
          <span>tempo de plantão</span>
        </div>
      {/if}
      <button type="button" class="btn" onclick={() => (tab = 'dicas')}>
        <svg width="16" height="16" viewBox="0 0 24 24" aria-hidden="true" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M9 18h6"></path><path d="M10 22h4"></path><path d="M12 2a7 7 0 0 0-4 12.7V16h8v-1.3A7 7 0 0 0 12 2z"></path></svg>
        Dica <span class="mono faint small">{hintsRevealed}/{hints.length}</span>
      </button>
      <button type="button" class="btn btn-primary" onclick={verify} disabled={checking}>
        {#if checking}<span class="spinner"></span>Verificando{:else}Verificar <span class="kbd">Ctrl+Enter</span>{/if}
      </button>
      {#if confirmStop}
        <button type="button" class="btn btn-danger" onclick={stop} disabled={stopping}>Confirmar: destruir o lab</button>
        <button type="button" class="btn btn-ghost" onclick={() => (confirmStop = false)}>Cancelar</button>
      {:else}
        <button type="button" class="btn btn-ghost" onclick={stop}>Encerrar</button>
      {/if}
    </div>
  </div>

  <main class="lab">
    <aside class="panel side">
      {#if result}
        <div class="result" role="status">
          <div class="result-head">
            <svg width="22" height="22" viewBox="0 0 24 24" aria-hidden="true"><circle cx="12" cy="12" r="11" fill="var(--green-soft)" stroke="var(--green)" stroke-width="1.5"></circle><polyline points="7.5 12.5 10.5 15.5 16.5 9" fill="none" stroke="var(--green)" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"></polyline></svg>
            <strong>Incidente resolvido</strong>
          </div>
          <div class="result-nums">
            <div><span class="mono">{duration(result.seconds)}</span><small>tempo</small></div>
            <div><span class="mono">{result.hints}</span><small>dicas</small></div>
            <div><span class="mono">{result.points ?? '—'}</span><small>pontos</small></div>
          </div>
          {#if result.solutionViewed}<p class="faint small">A solução foi aberta antes da correção, então esta tentativa não pontua.</p>{/if}
          <div class="result-actions">
            <button type="button" class="btn btn-primary" onclick={() => loadSolution(true)}>Ver solução e causa raiz</button>
            <a class="btn" href="#/">Voltar à central</a>
          </div>
        </div>
      {/if}

      <div class="tabs" role="tablist" aria-label="Conteúdo do incidente">
        {#each tabs as [tid, label] (tid)}
          <button type="button" role="tab" aria-selected={tab === tid} class:sel={tab === tid} onclick={() => (tab = tid)}>{label}</button>
        {/each}
      </div>

      <div class="tab-body">
        {#if tab === 'briefing'}
          <Markdown source={detail.briefing} />
        {:else if tab === 'guiado'}
          <ol class="steps">
            {#each steps as s, i (i)}
              {@const current = i === currentStep && !resolved}
              <li>
                <div class="step-icon">
                  {#if s.done}
                    <svg width="24" height="24" viewBox="0 0 24 24" role="img" aria-label="Concluído"><circle cx="12" cy="12" r="11" fill="var(--green-soft)" stroke="var(--green)" stroke-width="1.5"></circle><polyline points="7.5 12.5 10.5 15.5 16.5 9" fill="none" stroke="var(--green)" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"></polyline></svg>
                  {:else if current}
                    <svg width="24" height="24" viewBox="0 0 24 24" role="img" aria-label="Em andamento"><circle cx="12" cy="12" r="11" fill="var(--accent-soft)" stroke="var(--accent)" stroke-width="1.5"></circle><circle cx="12" cy="12" r="4" fill="var(--accent)"></circle></svg>
                  {:else if !s.hasVerify}
                    <svg width="24" height="24" viewBox="0 0 24 24" role="img" aria-label="Investigação livre"><circle cx="12" cy="12" r="10.5" fill="none" stroke="#3A4659" stroke-width="1.5"></circle><circle cx="12" cy="12" r="2" fill="#5B6778"></circle></svg>
                  {:else}
                    <svg width="24" height="24" viewBox="0 0 24 24" role="img" aria-label="Pendente"><circle cx="12" cy="12" r="10.5" fill="none" stroke="#3A4659" stroke-width="1.5" stroke-dasharray="3 3"></circle></svg>
                  {/if}
                </div>
                <div class="step-txt">
                  <div class="step-title">
                    <span class:dim={!s.done && !current}>{s.title}</span>
                    {#if s.optional}<span class="tag-bonus mono">bônus</span>{/if}
                  </div>
                  <span class="muted step-desc"><Markdown source={s.text} inline /></span>
                  {#if current}<span class="mono small accent">verificando automaticamente…</span>{/if}
                  {#if !s.hasVerify && !s.done}<span class="faint small">investigação: sem verificação automática</span>{/if}
                </div>
              </li>
            {/each}
          </ol>
        {:else if tab === 'conceitos'}
          <Markdown source={detail.concepts} />
        {:else if tab === 'dicas'}
          <div class="hints">
            {#each hints as h (h.n)}
              {@const canReveal = !h.text && h.n === hintsRevealed + 1}
              <div class="hint" class:shown={!!h.text}>
                <div class="hint-head">
                  <strong>Dica {h.n}</strong>
                  <span class="mono faint small">−{h.cost} pts</span>
                </div>
                {#if h.text}
                  <Markdown source={h.text} />
                {:else if canReveal}
                  <button type="button" class="btn reveal" onclick={() => reveal(h.n)} disabled={hintBusy}>Revelar dica {h.n}</button>
                {:else}
                  <span class="faint small">Revele a dica anterior primeiro.</span>
                {/if}
              </div>
            {/each}
            <div class="sol-box">
              {#if confirmSolution}
                <p>Abrir a solução agora <strong>zera a pontuação</strong> desta tentativa. Quer mesmo ver?</p>
                <div class="row-btns">
                  <button type="button" class="btn btn-danger" onclick={() => loadSolution(true)}>Ver a solução</button>
                  <button type="button" class="btn btn-ghost" onclick={() => (confirmSolution = false)}>Continuar tentando</button>
                </div>
              {:else}
                <button type="button" class="btn btn-ghost" onclick={() => loadSolution()}>Desistir e ver a solução</button>
              {/if}
            </div>
          </div>
        {:else if tab === 'solucao' && solution}
          <Markdown source={solution.solution} />
          {#if solution.rootCause}<hr class="sep" /><Markdown source={solution.rootCause} />{/if}
          {#if solution.references}<hr class="sep" /><Markdown source={solution.references} />{/if}
        {/if}
      </div>
    </aside>

    <section class="work">
      <div class="term-wrap">
        <Terminal scenarioId={id} label={`${detail.user}@${detail.hostname} · bash`} runtime={status?.runtime} />
      </div>

      <div class="bottom">
        <div class="panel mon">
          <div class="mon-head">
            <h2>Monitoramento ao vivo</h2>
            <span class="mono faint small">{detail.hostname} · 5 s</span>
          </div>
          {#if monitors.length}
            <div class="mon-grid">
              {#each monitors as m (m.label)}
                <div class="mcard">
                  <span class="muted small">{m.label}</span>
                  <span class="mono mval" style:color={m.color}>{m.value}</span>
                  <Sparkline values={m.series} max={m.max} color={m.color} />
                  <span class="faint small">{m.sub}</span>
                </div>
              {/each}
            </div>
          {:else}
            <p class="faint small">Coletando métricas do lab…</p>
          {/if}
        </div>

        <div class="panel last">
          <div class="mon-head">
            <h2>Última verificação</h2>
            <span class="mono faint small">{check ? new Date(check.at).toLocaleTimeString('pt-BR') : 'nenhuma ainda'}</span>
          </div>
          {#if check}
            <ul class="checks">
              {#each check.items as c, i (i)}
                <li>
                  {#if c.ok}
                    <svg width="18" height="18" viewBox="0 0 24 24" role="img" aria-label="Passou"><circle cx="12" cy="12" r="11" fill="var(--green-soft)"></circle><polyline points="7.5 12.5 10.5 15.5 16.5 9" fill="none" stroke="var(--green)" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round"></polyline></svg>
                  {:else}
                    <svg width="18" height="18" viewBox="0 0 24 24" role="img" aria-label="Falhou"><circle cx="12" cy="12" r="11" fill="#2A1216"></circle><line x1="8.5" y1="8.5" x2="15.5" y2="15.5" stroke="var(--red)" stroke-width="2.4" stroke-linecap="round"></line><line x1="15.5" y1="8.5" x2="8.5" y2="15.5" stroke="var(--red)" stroke-width="2.4" stroke-linecap="round"></line></svg>
                  {/if}
                  <span>{c.text}</span>
                </li>
              {/each}
            </ul>
            {#if !check.passed}<p class="faint small">Ainda não resolvido. Continue investigando ou peça uma dica.</p>{/if}
          {:else}
            <p class="faint small">Quando achar que resolveu, clique em <strong>Verificar</strong>. A verificação confere a causa raiz, não só o sintoma.</p>
          {/if}
        </div>
      </div>
    </section>
  </main>
{/if}

<style>
  .small { font-size: 12px; }
  .accent { color: var(--accent); }
  .green { color: var(--green); }
  .center { min-height: 60vh; display: flex; align-items: center; justify-content: center; padding: 24px; }
  .box { padding: 28px; display: flex; flex-direction: column; gap: 12px; align-items: flex-start; max-width: 520px; }
  .box h1 { font-size: 20px; }
  .box p { margin: 0; }

  .back {
    display: inline-flex; align-items: center; gap: 6px; color: var(--muted); text-decoration: none;
    font-size: 14px; padding: 8px 4px; min-height: 40px;
  }
  .back:hover { color: var(--text); }

  /* antes de assumir */
  .pre { max-width: 1200px; margin: 0 auto; padding: 20px 32px 56px; display: flex; flex-direction: column; gap: 12px; }
  .pre-grid { display: flex; flex-wrap: wrap; gap: 20px; align-items: flex-start; }
  .pre-main { flex: 999 1 560px; min-width: 0; padding: 28px; display: flex; flex-direction: column; gap: 14px; border-radius: 14px; }
  .pre-head { display: flex; align-items: center; gap: 10px; }
  .pre-main > h1 { font-size: 28px; letter-spacing: -0.02em; }
  .obj-title { font-size: 15px; margin-top: 6px; }
  .obj { margin: 0; padding-left: 20px; color: var(--text-2); display: flex; flex-direction: column; gap: 4px; }
  .pre-side { flex: 1 1 300px; min-width: 0; padding: 22px; display: flex; flex-direction: column; gap: 16px; border-radius: 14px; position: sticky; top: 16px; }
  dl { margin: 0; display: flex; flex-direction: column; gap: 10px; }
  dl div { display: flex; justify-content: space-between; gap: 12px; border-bottom: 1px solid #1C2532; padding-bottom: 10px; }
  dt { color: var(--muted); font-size: 14px; }
  dd { margin: 0; font-size: 13px; }
  .wide { width: 100%; min-height: 48px; font-size: 16px; }
  .pre-side p { margin: 0; }
  .starting { display: flex; gap: 12px; align-items: flex-start; color: var(--accent); }
  .starting div { display: flex; flex-direction: column; gap: 2px; color: var(--text); }
  .start-err { display: flex; flex-direction: column; gap: 6px; padding: 12px 14px; border-radius: 8px; background: var(--red-soft); border: 1px solid var(--red-line); color: #F9A9AD; font-size: 14px; }
  .start-err pre { margin: 0; max-height: 220px; overflow: auto; font-size: 12px; color: var(--text-2); white-space: pre-wrap; }

  /* barra do incidente */
  .incident {
    display: flex; flex-wrap: wrap; align-items: center; gap: 12px 18px; padding: 12px 24px;
    background: var(--red-soft); border-bottom: 1px solid var(--red-line); transition: background .4s, border-color .4s;
  }
  .incident.ok { background: #0D1A14; border-bottom-color: var(--green-line); }
  .pill.pulse { animation: pulse 1.8s ease-out infinite; }
  .inc-title { display: flex; flex-direction: column; min-width: 0; }
  .inc-title h1 { font-size: 18px; line-height: 1.3; font-weight: 600; }
  .inc-actions { margin-left: auto; display: flex; flex-wrap: wrap; align-items: center; gap: 10px; }
  .timer { display: flex; flex-direction: column; align-items: flex-end; padding-right: 6px; }
  .timer .mono { font-size: 24px; font-weight: 700; letter-spacing: .02em; line-height: 1.1; }
  .timer span:last-child { font-size: 11px; color: var(--faint); text-transform: uppercase; letter-spacing: .08em; }

  /* lab */
  .lab { display: flex; flex-wrap: wrap; gap: 16px; padding: 16px; align-items: stretch; }
  .side { flex: 1 1 360px; min-width: 0; max-height: calc(100vh - 150px); display: flex; flex-direction: column; overflow: hidden; }
  .work { flex: 999 1 600px; min-width: 0; display: flex; flex-direction: column; gap: 16px; }
  .term-wrap { height: max(420px, calc(100vh - 470px)); }

  .result { margin: 14px 14px 0; padding: 16px; border-radius: 10px; background: #0D1A14; border: 1px solid var(--green-line); display: flex; flex-direction: column; gap: 12px; }
  .result-head { display: flex; align-items: center; gap: 10px; }
  .result-nums { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 8px; }
  .result-nums div { display: flex; flex-direction: column; }
  .result-nums .mono { font-size: 20px; font-weight: 700; }
  .result-nums small { color: var(--faint); font-size: 12px; }
  .result p { margin: 0; }
  .result-actions { display: flex; flex-wrap: wrap; gap: 8px; }

  .tabs { display: flex; border-bottom: 1px solid var(--line); overflow-x: auto; flex: none; }
  .tabs button {
    flex: 1 0 auto; min-height: 46px; padding: 0 14px; border: none; border-bottom: 2px solid transparent;
    background: transparent; color: var(--muted); font-size: 14px; font-weight: 600; cursor: pointer;
  }
  .tabs button:hover { color: var(--text); }
  .tabs button.sel { color: var(--text); border-bottom-color: var(--accent); }
  .tab-body { padding: 20px; overflow-y: auto; flex: 1; min-height: 0; }

  .steps { list-style: none; margin: 0; padding: 0; }
  .steps li { display: flex; gap: 14px; padding: 12px 0; border-bottom: 1px solid #18202B; }
  .steps li:last-child { border-bottom: none; }
  .step-icon { flex: 0 0 26px; display: flex; justify-content: center; padding-top: 1px; }
  .step-txt { display: flex; flex-direction: column; gap: 4px; min-width: 0; }
  .step-title { display: flex; flex-wrap: wrap; align-items: center; gap: 8px; font-weight: 600; }
  .step-title .dim { color: var(--muted); }
  .step-desc { font-size: 14px; }
  .tag-bonus { font-size: 11px; color: var(--blue); border: 1px solid #22415C; border-radius: 4px; padding: 0 6px; font-weight: 400; }

  .hints { display: flex; flex-direction: column; gap: 12px; }
  .hint { border: 1px solid #222C3A; border-radius: 10px; padding: 14px 16px; display: flex; flex-direction: column; gap: 8px; }
  .hint.shown { border-color: var(--accent-line); background: #0D1F1E; }
  .hint :global(.md p:last-child) { margin-bottom: 0; }
  .hint-head { display: flex; align-items: center; justify-content: space-between; }
  .reveal { align-self: flex-start; min-height: 40px; border-color: var(--accent-line); background: var(--accent-soft); color: #7CE9DC; font-weight: 600; }
  .sol-box { margin-top: 8px; padding-top: 14px; border-top: 1px solid var(--line); display: flex; flex-direction: column; gap: 10px; }
  .sol-box p { margin: 0; color: var(--text-2); }
  .row-btns { display: flex; flex-wrap: wrap; gap: 8px; }
  .sep { border: none; border-top: 1px solid var(--line); margin: 22px 0; }

  .bottom { display: flex; flex-wrap: wrap; gap: 16px; }
  .mon { flex: 999 1 460px; min-width: 0; padding: 16px 18px; display: flex; flex-direction: column; gap: 12px; }
  .last { flex: 1 1 300px; min-width: 0; padding: 16px 18px; display: flex; flex-direction: column; gap: 10px; }
  .mon-head { display: flex; align-items: center; gap: 8px; }
  .mon-head h2 { font-size: 15px; }
  .mon-head .mono { margin-left: auto; }
  .mon p, .last p { margin: 0; }
  .mon-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(118px, 1fr)); gap: 10px; }
  .mcard { background: var(--panel-2); border: 1px solid #1C2532; border-radius: 10px; padding: 12px; display: flex; flex-direction: column; gap: 4px; min-width: 0; }
  .mval { font-size: 20px; font-weight: 700; }
  .checks { list-style: none; margin: 0; padding: 0; display: flex; flex-direction: column; gap: 8px; }
  .checks li { display: flex; gap: 10px; align-items: flex-start; font-size: 14px; color: var(--text-2); }
  .checks svg { flex: 0 0 18px; margin-top: 2px; }

  @media (max-width: 900px) {
    .side { max-height: none; }
    .incident { padding: 12px 16px; }
    .inc-actions { margin-left: 0; }
    .pre { padding: 16px 16px 40px; }
    .pre-side { position: static; }
  }
</style>
