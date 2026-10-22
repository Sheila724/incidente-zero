<script>
  import { onMount } from 'svelte';
  import TopBar from '../lib/TopBar.svelte';
  import Markdown from '../lib/Markdown.svelte';
  import { api, postmortemURL } from '../lib/api.js';
  import { sev, duration } from '../lib/format.js';

  let { id } = $props();

  // Fuso do navegador: os horários do Markdown saem no horário de quem escreve
  const TZ = Intl.DateTimeFormat().resolvedOptions().timeZone || '';

  let status = $state(null);
  let data = $state(null);
  let error = $state('');
  let saving = $state(false);
  let savedAt = $state(null);
  let dirty = $state(false);
  let view = $state('editar'); // editar | previa
  let copied = $state(false);
  let reference = $state('');
  let showReference = $state(false);

  let form = $state({ resumo: '', causa: '', evidencia: '', coincidencia: '', acoes: '' });

  const FIELDS = [
    {
      key: 'resumo', label: 'Resumo', rows: 3,
      help: 'O que aconteceu, qual foi o impacto e como foi resolvido, em duas ou três frases.',
      placeholder: 'Ex.: o checkout ficou fora do ar por 12 minutos porque...'
    },
    {
      key: 'causa', label: 'Causa raiz', rows: 4,
      help: 'Por que aconteceu? Vá além do gatilho: os 5 porquês ajudam a chegar no processo que falhou.',
      placeholder: 'Ex.: o binlog foi ligado para um teste e ficou sem expiração...'
    },
    {
      key: 'evidencia', label: 'Evidência que confirmou a causa', rows: 3,
      help: 'Qual log, comando ou métrica provou a causa? Algo que outra pessoa possa conferir.',
      placeholder: 'Ex.: journalctl -u mariadb mostrando "Disk is full writing..."'
    },
    {
      key: 'coincidencia', label: 'O que só coincidiu com a recuperação', rows: 3,
      help: 'O que mudou junto com a melhora, mas não explica a causa (um restart, um horário, um deploy próximo).',
      placeholder: 'Ex.: reiniciar o serviço fez o erro sumir por alguns minutos...'
    },
    {
      key: 'acoes', label: 'Ações de prevenção', rows: 4,
      help: 'Uma ação por linha. Prefira ações que detectem cedo ou impeçam de acontecer de novo.',
      placeholder: '- alerta de disco em 80%\n- expiração do binlog em 1 dia'
    }
  ];

  async function load() {
    try {
      const [st, pm] = await Promise.all([api.status(), api.postmortem(id, TZ)]);
      status = st;
      data = pm;
      form = { ...form, ...pickFields(pm.postmortem) };
      savedAt = pm.saved ? pm.postmortem.savedAt : null;
      dirty = false;
      error = '';
    } catch (e) {
      error = e.message;
    }
  }

  function pickFields(pm) {
    const out = {};
    for (const f of FIELDS) out[f.key] = pm?.[f.key] ?? '';
    return out;
  }

  async function save() {
    if (saving || !data?.resolved) return;
    saving = true;
    try {
      data = await api.savePostmortem(id, $state.snapshot(form), TZ);
      savedAt = data.postmortem.savedAt;
      dirty = false;
      error = '';
    } catch (e) {
      error = e.message;
    } finally {
      saving = false;
    }
  }

  async function copy() {
    if (dirty) await save();
    try {
      await navigator.clipboard.writeText(data.markdown);
      copied = true;
      setTimeout(() => (copied = false), 2000);
    } catch {
      error = 'Não consegui copiar: use "Baixar .md".';
    }
  }

  async function toggleReference() {
    showReference = !showReference;
    if (showReference && !reference) {
      try {
        reference = (await api.solution(id)).rootCause;
      } catch (e) {
        error = e.message;
      }
    }
  }

  function onKey(e) {
    if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 's') {
      e.preventDefault();
      save();
    }
  }

  function onLeave(e) {
    if (dirty) {
      e.preventDefault();
      e.returnValue = '';
    }
  }

  onMount(load);

  const sc = $derived(data?.scenario);
  const sv = $derived(sev(sc?.severity));
  const filled = $derived(FIELDS.filter((f) => form[f.key].trim()).length);
  const start = $derived(data?.timeline?.[0] ? new Date(data.timeline[0].at).getTime() : 0);

  function hora(t) {
    return new Date(t).toLocaleTimeString('pt-BR', { hour: '2-digit', minute: '2-digit', second: '2-digit' });
  }
  function desde(t) {
    const s = Math.max(0, Math.round((new Date(t).getTime() - start) / 1000));
    return `+${String(Math.floor(s / 60)).padStart(2, '0')}:${String(s % 60).padStart(2, '0')}`;
  }
  const KIND = {
    start: { label: 'Alerta', color: 'var(--red)' },
    hint: { label: 'Dica', color: 'var(--violet)' },
    check: { label: 'Verificação', color: 'var(--blue)' },
    solution: { label: 'Solução', color: 'var(--pink)' },
    resolved: { label: 'Resolvido', color: 'var(--green)' }
  };
  function descricao(e) {
    return e.kind === 'start' ? `Plantão assumido: ${sc?.alertTitle}` : e.detail;
  }
</script>

<svelte:window onkeydown={onKey} onbeforeunload={onLeave} />

<TopBar {status} current="postmortem" />

<main class="pm">
  {#if error}<p class="err" role="alert">{error}</p>{/if}

  {#if !data}
    {#if !error}<p class="muted">Carregando…</p>{/if}
  {:else}
    <header class="head">
      <div class="crumbs">
        <a href="#/">Central</a>
        <span class="faint">/</span>
        <a href={`#/lab/${id}`}>#{sc.number} {sc.title}</a>
      </div>
      <div class="eyebrow mono">Postmortem sem culpados</div>
      <h1>{sc.title}</h1>
      <div class="chips">
        <span class="chip"><span class="dot" style:background={sv.color}></span>{sv.label}</span>
        <span class="chip mono">{data.hostname}</span>
        {#if data.resolved}
          <span class="chip">MTTR <strong class="mono">{duration(data.seconds)}</strong></span>
          <span class="chip">Dicas <strong class="mono">{data.hints}</strong></span>
          <span class="chip">Pontos <strong class="mono">{data.points}</strong></span>
        {/if}
      </div>
    </header>

    {#if !data.resolved}
      <section class="panel empty">
        <h2>Resolva o incidente primeiro</h2>
        <p class="muted">O postmortem é escrito depois da resolução, com a linha do tempo do seu plantão.</p>
        <a class="btn btn-primary" href={`#/lab/${id}`}>Ir para o lab</a>
      </section>
    {:else}
      <div class="grid">
        <aside class="panel timeline">
          <h2>Linha do tempo</h2>
          <p class="faint small">Registrada automaticamente durante o plantão.</p>
          {#if data.timeline.length}
            <ol>
              {#each data.timeline as e, i (i)}
                {@const k = KIND[e.kind] ?? { label: e.kind, color: 'var(--muted)' }}
                <li>
                  <span class="mark" style:background={k.color}></span>
                  <div>
                    <div class="when mono"><span>{hora(e.at)}</span><span class="faint">{desde(e.at)}</span></div>
                    <div class="what" title={k.label}>{descricao(e)}</div>
                  </div>
                </li>
              {/each}
            </ol>
          {:else}
            <p class="muted small">Sem eventos registrados nesta resolução.</p>
          {/if}
        </aside>

        <section class="panel editor">
          <div class="editor-head">
            <div class="seg" role="tablist" aria-label="Modo">
              <button type="button" role="tab" aria-selected={view === 'editar'} class:sel={view === 'editar'} onclick={() => (view = 'editar')}>Escrever</button>
              <button type="button" role="tab" aria-selected={view === 'previa'} class:sel={view === 'previa'} onclick={async () => { if (dirty) await save(); view = 'previa'; }}>Prévia</button>
            </div>
            <span class="faint small mono">{filled}/{FIELDS.length} seções</span>
          </div>

          {#if view === 'editar'}
            <form onsubmit={(e) => { e.preventDefault(); save(); }}>
              {#each FIELDS as f (f.key)}
                <label class="field">
                  <span class="lbl">{f.label}</span>
                  <span class="help">{f.help}</span>
                  <textarea
                    rows={f.rows}
                    placeholder={f.placeholder}
                    bind:value={form[f.key]}
                    oninput={() => (dirty = true)}
                  ></textarea>
                </label>
              {/each}
            </form>
          {:else}
            <div class="preview"><Markdown source={data.markdown} /></div>
          {/if}

          <div class="actions">
            <button type="button" class="btn btn-primary" onclick={save} disabled={saving || !dirty}>
              {saving ? 'Salvando…' : 'Salvar'} <span class="kbd">Ctrl+S</span>
            </button>
            <a class="btn" href={postmortemURL(id, TZ)} download>Baixar .md</a>
            <button type="button" class="btn" onclick={copy}>{copied ? 'Copiado!' : 'Copiar Markdown'}</button>
            <span class="state small" class:warn={dirty}>
              {#if dirty}Alterações não salvas{:else if savedAt}Salvo às {hora(savedAt)}{/if}
            </span>
          </div>
        </section>
      </div>

      <section class="panel ref">
        <button type="button" class="ref-toggle" aria-expanded={showReference} onclick={toggleReference}>
          <span>Comparar com o postmortem de referência</span>
          <span class="faint small">{showReference ? 'Ocultar' : savedAt ? 'Mostrar' : 'Escreva o seu primeiro'}</span>
        </button>
        {#if showReference}
          <div class="ref-body"><Markdown source={reference || 'Carregando…'} /></div>
        {/if}
      </section>
    {/if}
  {/if}
</main>

<style>
  .pm { max-width: 1360px; margin: 0 auto; padding: 28px 32px 56px; display: flex; flex-direction: column; gap: 22px; }
  .small { font-size: 13px; }
  .err { margin: 0; padding: 12px 14px; border-radius: 10px; background: var(--red-soft); border: 1px solid var(--red-line); color: #F9A9AD; }

  .head { display: flex; flex-direction: column; gap: 8px; }
  .crumbs { display: flex; gap: 8px; font-size: 14px; }
  .crumbs a { text-decoration: none; color: var(--muted); }
  .crumbs a:hover { color: var(--text); }
  .eyebrow { font-size: 12px; letter-spacing: .12em; color: var(--accent); text-transform: uppercase; margin-top: 6px; }
  h1 { font-size: 30px; line-height: 1.2; letter-spacing: -0.02em; }
  .chips { display: flex; flex-wrap: wrap; gap: 8px; margin-top: 4px; }
  .chip { display: inline-flex; align-items: center; gap: 7px; padding: 5px 11px; border-radius: 999px; background: var(--chip); border: 1px solid var(--line); font-size: 13px; color: var(--text-2); }
  .chip strong { color: var(--text); }
  .dot { width: 8px; height: 8px; border-radius: 50%; display: inline-block; }

  .empty { padding: 28px; display: flex; flex-direction: column; gap: 10px; align-items: flex-start; }
  .empty p { margin: 0; }

  .grid { display: grid; grid-template-columns: minmax(280px, 380px) 1fr; gap: 18px; align-items: start; }
  @media (max-width: 960px) { .grid { grid-template-columns: 1fr; } }

  .timeline { padding: 20px; position: sticky; top: 16px; }
  .timeline h2, .editor h2 { font-size: 17px; }
  .timeline > p { margin: 4px 0 14px; }
  ol { list-style: none; margin: 0; padding: 0; display: flex; flex-direction: column; }
  li { display: grid; grid-template-columns: 14px 1fr; gap: 12px; padding: 0 0 16px; position: relative; }
  li:not(:last-child)::before { content: ''; position: absolute; left: 4px; top: 14px; bottom: 0; width: 2px; background: var(--line); }
  .mark { width: 10px; height: 10px; border-radius: 50%; margin-top: 6px; box-shadow: 0 0 0 3px var(--panel); }
  .when { display: flex; gap: 10px; font-size: 12px; color: var(--text-2); }
  .what { font-size: 14px; color: var(--text-2); }
  .what strong { font-weight: 600; margin-right: 4px; }

  .editor { padding: 20px; display: flex; flex-direction: column; gap: 16px; }
  .editor-head { display: flex; align-items: center; justify-content: space-between; gap: 12px; }
  .seg { display: inline-flex; padding: 3px; border-radius: 9px; background: var(--panel-2); border: 1px solid var(--line); }
  .seg button { border: 0; background: transparent; color: var(--muted); padding: 6px 14px; border-radius: 7px; cursor: pointer; font-weight: 500; }
  .seg button.sel { background: var(--chip); color: var(--text); }

  form { display: flex; flex-direction: column; gap: 18px; }
  .field { display: flex; flex-direction: column; gap: 4px; }
  .lbl { font-weight: 600; }
  .help { font-size: 13px; color: var(--muted); }
  textarea {
    margin-top: 4px; width: 100%; resize: vertical; min-height: 64px;
    padding: 10px 12px; border-radius: 9px; border: 1px solid var(--line-2);
    background: var(--panel-2); color: var(--text); font: inherit; font-size: 14px; line-height: 1.55;
  }
  textarea::placeholder { color: #5D6B7D; }
  textarea:focus { outline: none; border-color: var(--accent-line); box-shadow: 0 0 0 3px var(--accent-soft); }

  .preview { padding: 4px 2px; }
  .actions { display: flex; flex-wrap: wrap; align-items: center; gap: 10px; padding-top: 6px; border-top: 1px solid var(--line); }
  .state { color: var(--faint); margin-left: auto; }
  .state.warn { color: var(--violet); }

  .ref { padding: 0; overflow: hidden; }
  .ref-toggle { width: 100%; display: flex; justify-content: space-between; align-items: center; gap: 12px; padding: 16px 20px; background: transparent; border: 0; color: var(--text); cursor: pointer; font-weight: 600; text-align: left; }
  .ref-toggle:hover { background: var(--panel-2); }
  .ref-body { padding: 4px 20px 20px; border-top: 1px solid var(--line); }

  @media (max-width: 640px) {
    .pm { padding: 20px 16px 40px; }
    h1 { font-size: 24px; }
    .state { margin-left: 0; width: 100%; }
  }
</style>
