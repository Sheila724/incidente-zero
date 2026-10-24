<script>
  import { onMount } from 'svelte';
  import TopBar from '../lib/TopBar.svelte';
  import { api } from '../lib/api.js';
  import { sev, duration, clock, LEVELS } from '../lib/format.js';
  import {
    plantao, atualizarPlantao, desbloquearAudio, pedirNotificacao, tocarAlarme,
    alternarMudo, RITMOS, DURACOES
  } from '../lib/plantao.svelte.js';

  let status = $state(null);
  let niveis = $state([]); // trilhas com cenários instalados
  let duracaoMin = $state(60);
  let ritmo = $state('normal');
  let trilhas = $state([]);
  let enviando = $state(false);
  let erro = $state('');
  let confirmarEncerrar = $state(false);
  let agora = $state(Date.now());
  let permissao = $state(typeof Notification !== 'undefined' ? Notification.permission : 'unsupported');

  onMount(() => {
    api.status().then((s) => (status = s)).catch(() => {});
    api.scenarios().then((l) => {
      const ids = new Set(l.scenarios.map((s) => s.level));
      niveis = LEVELS.filter((n) => ids.has(n.id)).map((n) => ({
        ...n, total: l.scenarios.filter((s) => s.level === n.id).length
      }));
      if (!trilhas.length) trilhas = niveis.map((n) => n.id);
    });
    atualizarPlantao();
    const t = setInterval(() => (agora = Date.now()), 1000);
    return () => clearInterval(t);
  });

  const turno = $derived(plantao.dados?.turno ?? null);
  const incidente = $derived(plantao.dados?.incidente ?? null);
  const ultimo = $derived(plantao.dados?.anteriores?.[0] ?? null);
  const historico = $derived(plantao.dados?.anteriores ?? []);
  const relogio = $derived(agora + plantao.diferenca);
  const restante = $derived(turno ? Math.max(0, (turno.fim - relogio) / 1000) : 0);
  const atendidos = $derived((turno?.incidentes ?? []).filter((i) => i.status !== 'falhou' && i.status !== 'preparando'));

  function hora(ms) {
    return new Date(ms).toLocaleTimeString('pt-BR', { hour: '2-digit', minute: '2-digit' });
  }
  function dataHora(ms) {
    return new Date(ms).toLocaleString('pt-BR', { day: '2-digit', month: '2-digit', hour: '2-digit', minute: '2-digit' });
  }
  function seg(a, b) {
    return a && b ? (new Date(b) - new Date(a)) / 1000 : null;
  }
  // Cursos e livros gratuitos por trilha, para quem travou
  const CURSOS = {
    iniciante: [
      { titulo: 'Introduction to Linux (LFS101)', url: 'https://training.linuxfoundation.org/training/introduction-to-linux/', desc: 'Curso gratuito da Linux Foundation (em inglês)' }
    ],
    intermediario: [
      { titulo: 'Webinars gratuitos do Zabbix', url: 'https://www.zabbix.com/webinars', desc: 'Vídeos oficiais sobre monitoramento com Zabbix' }
    ],
    avancado: [
      { titulo: 'Descomplicando o Kubernetes', url: 'https://github.com/badtuxx/DescomplicandoKubernetes', desc: 'Livro gratuito da LINUXtips, em português' },
      { titulo: 'Introduction to Kubernetes (LFS158)', url: 'https://training.linuxfoundation.org/training/introduction-to-kubernetes/', desc: 'Curso gratuito da Linux Foundation (em inglês)' },
      { titulo: 'Kubernetes Basics', url: 'https://kubernetes.io/docs/tutorials/kubernetes-basics/', desc: 'Tutorial interativo da documentação oficial' }
    ],
    todos: [
      { titulo: 'Livros de SRE do Google', url: 'https://sre.google/books/', desc: 'Site Reliability Engineering e The Site Reliability Workbook, gratuitos online (em inglês)' }
    ]
  };
  function cursosPara(lista) {
    const niveis = [...new Set(lista.map((i) => i.level).filter(Boolean))];
    return [...niveis.flatMap((n) => CURSOS[n] ?? []), ...CURSOS.todos];
  }

  const NOME_STATUS = {
    alertando: 'Alertando', reconhecido: 'Em atendimento', resolvido: 'Resolvido',
    'nao-resolvido': 'Não resolvido'
  };

  function alternarTrilha(id) {
    trilhas = trilhas.includes(id) ? trilhas.filter((t) => t !== id) : [...trilhas, id];
  }

  async function assumir() {
    desbloquearAudio();
    await pedirNotificacao();
    permissao = typeof Notification !== 'undefined' ? Notification.permission : 'unsupported';
    enviando = true;
    erro = '';
    try {
      await api.assumirPlantao({ duracaoMin, ritmo, trilhas });
      await atualizarPlantao();
    } catch (e) {
      erro = e.message;
    } finally {
      enviando = false;
    }
  }

  async function encerrar() {
    if (!confirmarEncerrar) {
      confirmarEncerrar = true;
      return;
    }
    enviando = true;
    try {
      await api.encerrarPlantao();
      await atualizarPlantao();
    } catch (e) {
      erro = e.message;
    } finally {
      enviando = false;
      confirmarEncerrar = false;
    }
  }

  function testarSom() {
    desbloquearAudio();
    setTimeout(tocarAlarme, 50);
  }
</script>

<TopBar {status} current="plantao" />

<main>
  <section class="hero">
    <div>
      <span class="mono eyebrow">Modo Plantão</span>
      <h1>{turno ? 'Você está de plantão' : 'Assuma um turno de plantão'}</h1>
      <p class="muted">
        Os incidentes caem sem aviso, em horários sorteados. Quando o alarme tocar, reconheça, investigue e resolva,
        como num plantão de verdade. Sem modo guiado.
      </p>
    </div>
  </section>

  {#if !plantao.carregado}
    <p class="muted">Carregando…</p>
  {:else if turno}
    <section class="row">
      <article class="panel turno">
        <div class="turno-head">
          <span class="dot" aria-hidden="true"></span>
          <div>
            <h2>De plantão até {hora(turno.fim)}</h2>
            <span class="mono faint">início {hora(turno.inicio)} · ritmo {turno.ritmo} · {turno.trilhas.map((t) => LEVELS.find((l) => l.id === t)?.name ?? t).join(', ')}</span>
          </div>
          <span class="mono restante" title="Tempo restante do turno">{clock(restante)}</span>
        </div>

        {#if incidente && incidente.status === 'reconhecido'}
          {@const sv = sev(incidente.severity)}
          <div class="atendendo">
            <span class="pill" style:background={sv.color} style:color={sv.ink}>{sv.pill}</span>
            <div>
              <strong>{incidente.alertTitle}</strong>
              <span class="mono faint">#{incidente.number} · em atendimento desde {hora(incidente.reconhecidoEm)}{incidente.escalado ? ' · escalado' : ''}</span>
            </div>
            <a class="btn btn-primary" href={`#/lab/${incidente.id}`}>Abrir o lab</a>
          </div>
        {:else if incidente && incidente.status === 'alertando'}
          <div class="calmo"><p>Há um incidente tocando. Reconheça no alerta.</p></div>
        {:else if restante <= 0}
          <div class="calmo"><p>O turno acabou. O relatório aparece em instantes.</p></div>
        {:else}
          <div class="calmo">
            <span class="radar" aria-hidden="true"></span>
            <p>Tudo calmo por enquanto. Deixe esta aba aberta: quando cair um incidente, o alarme toca e aparece um alerta.</p>
          </div>
        {/if}

        <div class="avisos">
          {#if permissao === 'denied'}
            <p class="faint small">As notificações do navegador estão bloqueadas para este site. O alerta continua aparecendo na tela, com som.</p>
          {:else if permissao === 'default'}
            <button type="button" class="btn btn-ghost" onclick={async () => { await pedirNotificacao(); permissao = Notification.permission; }}>Permitir notificações</button>
          {/if}
          <button type="button" class="btn btn-ghost" onclick={testarSom}>Testar o alarme</button>
          <button type="button" class="btn btn-ghost" onclick={alternarMudo} aria-pressed={plantao.mudo}>{plantao.mudo ? 'Som desligado' : 'Som ligado'}</button>
          <span class="espaco"></span>
          {#if confirmarEncerrar}
            <button type="button" class="btn btn-danger" onclick={encerrar} disabled={enviando}>Confirmar: encerrar o turno</button>
            <button type="button" class="btn btn-ghost" onclick={() => (confirmarEncerrar = false)}>Cancelar</button>
          {:else}
            <button type="button" class="btn btn-ghost" onclick={encerrar}>Encerrar turno</button>
          {/if}
        </div>
        {#if erro}<p class="erro">{erro}</p>{/if}
      </article>

      <article class="panel numeros">
        <h2>Neste turno</h2>
        <dl>
          <div><dt>Incidentes</dt><dd class="mono">{turno.resumo.incidentes}</dd></div>
          <div><dt>Resolvidos</dt><dd class="mono">{turno.resumo.resolvidos}</dd></div>
          <div><dt>MTTA</dt><dd class="mono">{turno.resumo.mttaSegundos ? duration(turno.resumo.mttaSegundos) : '—'}</dd></div>
          <div><dt>MTTR</dt><dd class="mono">{turno.resumo.mttrSegundos ? duration(turno.resumo.mttrSegundos) : '—'}</dd></div>
          <div><dt>Escalados</dt><dd class="mono">{turno.resumo.escalados}</dd></div>
          <div><dt>Pontos</dt><dd class="mono">{turno.resumo.pontos}</dd></div>
        </dl>
      </article>
    </section>

    {#if atendidos.length}
      {@render tabela(atendidos, 'Incidentes do turno')}
    {/if}
  {:else}
    {#if ultimo}
      {@render relatorio(ultimo)}
    {/if}

    <article class="panel form">
      <h2>{ultimo ? 'Assumir outro turno' : 'Novo turno'}</h2>

      <fieldset>
        <legend>Duração</legend>
        <div class="opcoes">
          {#each DURACOES as d (d.min)}
            <label class="opcao" class:sel={duracaoMin === d.min}>
              <input type="radio" name="duracao" value={d.min} bind:group={duracaoMin} />
              <span class="mono">{d.nome}</span>
            </label>
          {/each}
        </div>
      </fieldset>

      <fieldset>
        <legend>Ritmo</legend>
        <div class="opcoes">
          {#each RITMOS as r (r.id)}
            <label class="opcao larga" class:sel={ritmo === r.id}>
              <input type="radio" name="ritmo" value={r.id} bind:group={ritmo} />
              <span><strong>{r.nome}</strong><small class="faint">{r.desc}</small></span>
            </label>
          {/each}
        </div>
      </fieldset>

      <fieldset>
        <legend>Trilhas</legend>
        <div class="opcoes">
          {#each niveis as n (n.id)}
            <label class="opcao larga" class:sel={trilhas.includes(n.id)}>
              <input type="checkbox" checked={trilhas.includes(n.id)} onchange={() => alternarTrilha(n.id)} />
              <span><strong>{n.name}</strong><small class="faint">{n.desc} · {n.total} cenários</small></span>
            </label>
          {/each}
        </div>
      </fieldset>

      <p class="faint small">
        Mantenha esta aba aberta durante o turno. Ao assumir, o navegador pede permissão para notificações: com ela, o alerta aparece mesmo com a aba em segundo plano.
        O primeiro incidente cai alguns minutos depois de assumir. Se ninguém reconhecer em 5 minutos, ele é escalado e perde 20 pontos.
      </p>
      {#if erro}<p class="erro">{erro}</p>{/if}
      <div class="form-acoes">
        <button type="button" class="btn btn-primary grande" onclick={assumir} disabled={enviando || !trilhas.length}>
          {enviando ? 'Assumindo…' : 'Assumir plantão'}
        </button>
        <button type="button" class="btn btn-ghost" onclick={testarSom}>Testar o alarme</button>
      </div>
    </article>

    {#if historico.length > 1}
      <section class="panel hist">
        <h2>Turnos anteriores</h2>
        <div class="tabela">
          <table>
            <thead><tr><th>Início</th><th>Ritmo</th><th>Incidentes</th><th>Resolvidos</th><th>MTTA</th><th>MTTR</th><th>Pontos</th></tr></thead>
            <tbody>
              {#each historico.slice(1) as t (t.id)}
                <tr>
                  <td class="mono">{dataHora(t.inicio)}</td>
                  <td>{t.ritmo}</td>
                  <td class="mono">{t.resumo.incidentes}</td>
                  <td class="mono">{t.resumo.resolvidos}</td>
                  <td class="mono">{t.resumo.mttaSegundos ? duration(t.resumo.mttaSegundos) : '—'}</td>
                  <td class="mono">{t.resumo.mttrSegundos ? duration(t.resumo.mttrSegundos) : '—'}</td>
                  <td class="mono">{t.resumo.pontos}</td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      </section>
    {/if}
  {/if}
</main>

{#snippet relatorio(t)}
  <article class="panel relatorio">
    <div class="rel-head">
      <div>
        <span class="mono eyebrow">Relatório do turno</span>
        <h2>{dataHora(t.inicio)} a {hora(t.encerradoEm || t.fim)}</h2>
        <span class="mono faint">ritmo {t.ritmo} · {t.trilhas.map((x) => LEVELS.find((l) => l.id === x)?.name ?? x).join(', ')}</span>
      </div>
    </div>
    <dl class="rel-nums">
      <div><dt>Incidentes</dt><dd class="mono">{t.resumo.incidentes}</dd></div>
      <div><dt>Resolvidos</dt><dd class="mono">{t.resumo.resolvidos}</dd></div>
      <div><dt>Não resolvidos</dt><dd class="mono">{t.resumo.naoResolvidos}</dd></div>
      <div><dt>Escalados</dt><dd class="mono">{t.resumo.escalados}</dd></div>
      <div><dt>MTTA</dt><dd class="mono">{t.resumo.mttaSegundos ? duration(t.resumo.mttaSegundos) : '—'}</dd></div>
      <div><dt>MTTR</dt><dd class="mono">{t.resumo.mttrSegundos ? duration(t.resumo.mttrSegundos) : '—'}</dd></div>
      <div><dt>Pontos</dt><dd class="mono">{t.resumo.pontos}</dd></div>
    </dl>
    {#if t.resumo.incidentes === 0}
      <p class="muted">Nenhum incidente caiu neste turno.</p>
    {/if}
  </article>
  {@const lista = t.incidentes.filter((i) => i.alertaEm)}
  {#if lista.length}
    {@render tabela(lista, 'Incidentes do turno')}
    {@render estudar(lista.filter((i) => i.estudo))}
  {/if}
{/snippet}

{#snippet estudar(travou)}
  <section class="panel estudo">
    <h2>Para estudar</h2>
    {#if !travou.length}
      <p class="muted">Você não travou em nenhum incidente deste turno: sem dicas, sem solução aberta e dentro do tempo estimado.</p>
    {:else}
      <p class="muted">Os incidentes em que você teve mais dificuldade, com material para revisar. Refazer no modo guiado também ajuda a fixar o caminho.</p>
      {#each travou as i (i.id + i.criadoEm)}
        {@const sv = sev(i.severity)}
        <article class="est-item">
          <div class="est-head">
            <strong><span class="sevdot" style:background={sv.color}></span>#{i.number} {i.title}</strong>
            <a class="btn btn-ghost refazer" href={`#/lab/${i.id}`}>Refazer no modo guiado</a>
          </div>
          <div class="motivos">
            {#each i.estudo.motivos as m (m)}<span class="motivo">{m}</span>{/each}
          </div>
          {#if i.estudo.links.length}
            <ul class="links">
              {#each i.estudo.links as l (l.url)}
                <li><a href={l.url} target="_blank" rel="noopener noreferrer">{l.titulo}</a>{#if l.desc}<span class="faint">&nbsp;· {l.desc}</span>{/if}</li>
              {/each}
            </ul>
          {/if}
        </article>
      {/each}
      <div class="cursos">
        <h3>Cursos e livros gratuitos</h3>
        <ul class="links">
          {#each cursosPara(travou) as c (c.url)}
            <li><a href={c.url} target="_blank" rel="noopener noreferrer">{c.titulo}</a><span class="faint">&nbsp;· {c.desc}</span></li>
          {/each}
        </ul>
      </div>
    {/if}
  </section>
{/snippet}

{#snippet tabela(lista, titulo)}
  <section class="panel hist">
    <h2>{titulo}</h2>
    <div class="tabela">
      <table>
        <thead><tr><th>Alerta</th><th>Incidente</th><th>Reconhecido em</th><th>Resolvido em</th><th>Situação</th><th>Pontos</th></tr></thead>
        <tbody>
          {#each lista as i (i.id + i.criadoEm)}
            {@const sv = sev(i.severity)}
            <tr>
              <td class="mono">{hora(i.alertaEm)}</td>
              <td><span class="sevdot" style:background={sv.color}></span>#{i.number} {i.title}</td>
              <td class="mono">{seg(i.alertaEm, i.reconhecidoEm) != null ? duration(seg(i.alertaEm, i.reconhecidoEm)) : '—'}</td>
              <td class="mono">{seg(i.alertaEm, i.resolvidoEm) != null ? duration(seg(i.alertaEm, i.resolvidoEm)) : '—'}</td>
              <td>
                <span class="sit" class:ok={i.status === 'resolvido'} class:ruim={i.status === 'nao-resolvido'}>{NOME_STATUS[i.status] ?? i.status}</span>
                {#if i.escalado}<span class="sit ruim">escalado</span>{/if}
              </td>
              <td class="mono">{i.status === 'resolvido' || i.status === 'nao-resolvido' ? i.pontos : '—'}</td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  </section>
{/snippet}

<style>
  main { max-width: 1360px; margin: 0 auto; padding: 36px 32px 56px; display: flex; flex-direction: column; gap: 24px; }
  .eyebrow { font-size: 12px; letter-spacing: .12em; color: var(--accent); text-transform: uppercase; }
  .hero h1 { font-size: 32px; line-height: 1.15; margin: 6px 0; letter-spacing: -0.02em; }
  .hero p { margin: 0; max-width: 720px; }
  h2 { font-size: 18px; margin: 0; }
  .small { font-size: 13px; }
  .erro { color: #F9A9AD; margin: 0; }

  .row { display: flex; flex-wrap: wrap; gap: 20px; align-items: stretch; }
  .turno { flex: 999 1 560px; min-width: 0; padding: 22px; display: flex; flex-direction: column; gap: 18px; border-radius: 14px; border-color: var(--accent-line); }
  .turno-head { display: flex; align-items: center; gap: 14px; }
  .turno-head h2 { font-size: 22px; }
  .dot { width: 12px; height: 12px; border-radius: 50%; background: var(--accent); box-shadow: 0 0 0 4px var(--accent-soft); flex: none; animation: vivo 2s ease-in-out infinite; }
  .restante { margin-left: auto; font-size: 26px; font-weight: 700; color: var(--accent); }
  .calmo { display: flex; align-items: center; gap: 14px; padding: 16px; border-radius: 10px; background: var(--panel-2); border: 1px solid var(--line); }
  .calmo p { margin: 0; color: var(--text-2); }
  .radar { width: 26px; height: 26px; border-radius: 50%; border: 2px solid var(--accent-line); position: relative; flex: none; }
  .radar::after { content: ''; position: absolute; inset: 6px; border-radius: 50%; background: var(--accent); opacity: .7; animation: vivo 2s ease-in-out infinite; }
  .atendendo { display: flex; flex-wrap: wrap; align-items: center; gap: 12px; padding: 14px 16px; border-radius: 10px; background: var(--red-soft); border: 1px solid var(--red-line); }
  .atendendo div { display: flex; flex-direction: column; gap: 2px; flex: 1 1 240px; }
  .atendendo .mono { font-size: 12px; }
  .pill { font-size: 11px; font-weight: 700; padding: 3px 8px; border-radius: 6px; font-family: var(--mono); }
  .avisos { display: flex; flex-wrap: wrap; gap: 8px; align-items: center; }
  .avisos p { margin: 0; flex-basis: 100%; }
  .espaco { flex: 1; }

  .numeros { flex: 1 1 300px; padding: 22px; border-radius: 14px; display: flex; flex-direction: column; gap: 14px; }
  dl { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 12px 16px; margin: 0; }
  dt { font-size: 13px; color: var(--muted); }
  dd { margin: 2px 0 0; font-size: 22px; font-weight: 700; }

  .relatorio { padding: 22px; border-radius: 14px; display: flex; flex-direction: column; gap: 16px; border-color: var(--accent-line); }
  .rel-head h2 { font-size: 22px; margin: 4px 0; }
  .rel-head .mono { font-size: 12px; }
  .rel-nums { grid-template-columns: repeat(auto-fit, minmax(120px, 1fr)); }
  .relatorio p { margin: 0; }

  .form { padding: 22px; border-radius: 14px; display: flex; flex-direction: column; gap: 18px; }
  fieldset { border: 0; padding: 0; margin: 0; display: flex; flex-direction: column; gap: 10px; }
  legend { font-weight: 600; margin-bottom: 8px; padding: 0; }
  .opcoes { display: flex; flex-wrap: wrap; gap: 8px; }
  .opcao { display: flex; align-items: center; gap: 10px; padding: 10px 14px; border-radius: 10px; border: 1px solid var(--line); background: var(--panel-2); cursor: pointer; min-height: 44px; }
  .opcao.larga { flex: 1 1 240px; }
  .opcao span { display: flex; flex-direction: column; gap: 2px; }
  .opcao small { font-size: 12px; }
  .opcao.sel { border-color: var(--accent); background: var(--accent-soft); }
  .opcao input { accent-color: var(--accent); width: 16px; height: 16px; margin: 0; }
  .opcao:focus-within { outline: 2px solid var(--accent); outline-offset: 2px; }
  .form p { margin: 0; }
  .form-acoes { display: flex; flex-wrap: wrap; gap: 10px; }
  .grande { min-height: 48px; padding: 0 22px; font-size: 16px; }

  .hist { padding: 22px; border-radius: 14px; display: flex; flex-direction: column; gap: 12px; }
  .tabela { overflow-x: auto; }
  table { width: 100%; border-collapse: collapse; font-size: 14px; }
  th { text-align: left; font-weight: 500; color: var(--muted); font-size: 12px; padding: 8px 10px; border-bottom: 1px solid var(--line); white-space: nowrap; }
  td { padding: 10px; border-bottom: 1px solid #161E29; white-space: nowrap; }
  .sevdot { display: inline-block; width: 8px; height: 8px; border-radius: 50%; margin-right: 8px; }
  .sit { font-size: 12px; padding: 2px 8px; border-radius: 999px; background: var(--chip); color: var(--text-2); margin-right: 4px; }
  .sit.ok { background: var(--green-soft); color: var(--green); }
  .sit.ruim { background: var(--red-soft); color: #F9A9AD; }

  .estudo { padding: 22px; border-radius: 14px; display: flex; flex-direction: column; gap: 14px; }
  .estudo > p { margin: 0; }
  .est-item { display: flex; flex-direction: column; gap: 10px; padding: 14px 16px; border-radius: 10px; background: var(--panel-2); border: 1px solid var(--line); }
  .est-head { display: flex; flex-wrap: wrap; align-items: center; gap: 8px; }
  .refazer { margin-left: auto; min-height: 36px; font-size: 14px; }
  .motivos { display: flex; flex-wrap: wrap; gap: 6px; }
  .motivo { font-size: 12px; padding: 3px 10px; border-radius: 999px; background: var(--violet-soft); color: #CFC2F5; border: 1px solid var(--violet-line); }
  .links { margin: 0; padding-left: 18px; display: flex; flex-direction: column; gap: 6px; font-size: 14px; }
  .links a { color: var(--blue); }
  .cursos h3 { font-size: 15px; margin: 4px 0 8px; }
  @keyframes vivo { 0%, 100% { opacity: 1; } 50% { opacity: .35; } }
  @media (prefers-reduced-motion: reduce) { .dot, .radar::after { animation: none; } }
  @media (max-width: 640px) {
    main { padding: 24px 16px 40px; }
    .hero h1 { font-size: 26px; }
    .restante { font-size: 20px; }
  }
</style>
