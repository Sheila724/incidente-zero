<script>
  import { onMount, onDestroy } from 'svelte';
  import { Terminal } from '@xterm/xterm';
  import { FitAddon } from '@xterm/addon-fit';
  import '@xterm/xterm/css/xterm.css';
  import { terminalURL } from './api.js';

  let { scenarioId, label = 'bash', runtime = '' } = $props();

  let el;
  let term;
  let fit;
  let ws;
  let ro;
  let status = $state('conectando'); // conectando | conectado | encerrado
  const enc = new TextEncoder();

  function sendResize() {
    if (ws?.readyState === WebSocket.OPEN && term) {
      ws.send(JSON.stringify({ type: 'resize', cols: term.cols, rows: term.rows }));
    }
  }

  function connect() {
    status = 'conectando';
    ws = new WebSocket(terminalURL(scenarioId));
    ws.binaryType = 'arraybuffer';
    ws.onopen = () => {
      status = 'conectado';
      fit.fit();
      sendResize();
      term.focus();
    };
    ws.onmessage = (ev) => {
      term.write(typeof ev.data === 'string' ? ev.data : new Uint8Array(ev.data));
    };
    ws.onclose = () => {
      status = 'encerrado';
    };
  }

  export function reconnect() {
    try { ws?.close(); } catch { /* já fechado */ }
    term.reset();
    connect();
  }

  export function focus() {
    term?.focus();
  }

  onMount(() => {
    term = new Terminal({
      fontFamily: "'JetBrains Mono', ui-monospace, monospace",
      fontSize: 13,
      lineHeight: 1.25,
      cursorBlink: true,
      scrollback: 5000,
      allowProposedApi: false,
      theme: {
        background: '#06090D',
        foreground: '#E6EDF3',
        cursor: '#E6EDF3',
        cursorAccent: '#06090D',
        selectionBackground: '#1D5E5799',
        black: '#0C1118', brightBlack: '#5B6778',
        red: '#F2545B', brightRed: '#FF7A86',
        green: '#3DD68C', brightGreen: '#5EEAA8',
        yellow: '#F38BA8', brightYellow: '#F7A8BF', // sem amarelo/laranja na paleta
        blue: '#60A5FA', brightBlue: '#93C5FD',
        magenta: '#A78BFA', brightMagenta: '#C4B5FD',
        cyan: '#2DD4BF', brightCyan: '#5EEAD4',
        white: '#C3CDD9', brightWhite: '#F1F5F9'
      }
    });
    fit = new FitAddon();
    term.loadAddon(fit);
    term.open(el);
    fit.fit();

    term.onData((d) => {
      if (ws?.readyState === WebSocket.OPEN) ws.send(enc.encode(d));
    });
    term.onResize(sendResize);

    ro = new ResizeObserver(() => {
      try { fit.fit(); } catch { /* terminal oculto */ }
    });
    ro.observe(el);

    connect();
  });

  onDestroy(() => {
    ro?.disconnect();
    try { ws?.close(); } catch { /* já fechado */ }
    term?.dispose();
  });
</script>

<div class="term">
  <div class="bar">
    <svg width="16" height="16" viewBox="0 0 24 24" aria-hidden="true" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="4 17 10 11 4 5"></polyline><line x1="12" y1="19" x2="20" y2="19"></line></svg>
    <span class="mono label">{label}</span>
    <span class="state" class:on={status === 'conectado'} class:off={status === 'encerrado'}>
      <span class="dot"></span>
      {status}{#if runtime && status === 'conectado'}{` · ${runtime}`}{/if}
    </span>
  </div>
  <div class="screen" bind:this={el}></div>
  {#if status === 'encerrado'}
    <div class="overlay">
      <p>A sessão do terminal foi encerrada.</p>
      <button type="button" class="btn btn-primary" onclick={reconnect}>Reconectar</button>
    </div>
  {/if}
</div>

<style>
  .term {
    position: relative; display: flex; flex-direction: column; min-height: 0; height: 100%;
    background: #06090D; border: 1px solid var(--line); border-radius: var(--radius); overflow: hidden;
  }
  .bar {
    display: flex; align-items: center; gap: 10px; padding: 10px 14px;
    background: var(--panel-2); border-bottom: 1px solid #18202B; color: var(--muted);
  }
  .label { font-size: 13px; color: var(--text-2); }
  .state { margin-left: auto; display: inline-flex; align-items: center; gap: 6px; font-size: 12px; color: var(--muted); }
  .state .dot { width: 7px; height: 7px; background: var(--muted); }
  .state.on .dot { background: var(--green); }
  .state.off .dot { background: var(--red); }
  .screen { flex: 1; min-height: 0; padding: 10px 6px 6px 12px; }
  .screen :global(.xterm) { height: 100%; }
  .screen :global(.xterm-viewport) { background: transparent !important; }
  .overlay {
    position: absolute; inset: 44px 0 0 0; display: flex; flex-direction: column; gap: 12px;
    align-items: center; justify-content: center; background: rgba(6, 9, 13, .82); color: var(--text-2);
  }
  .overlay p { margin: 0; }
</style>
