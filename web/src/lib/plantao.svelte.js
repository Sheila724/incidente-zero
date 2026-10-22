// Estado global do Modo Plantão: consulta o servidor, toca o alarme e mostra a
// notificação do navegador quando um incidente cai.
import { api } from './api.js';

const CHAVE_MUDO = 'iz-plantao-mudo';

function lerMudo() {
  try {
    return localStorage.getItem(CHAVE_MUDO) === '1';
  } catch {
    return false;
  }
}

export const plantao = $state({
  dados: null, // resposta de GET /api/plantao
  diferenca: 0, // relógio do servidor − relógio local (ms)
  mudo: lerMudo(),
  carregado: false
});

let audio = null;
let alarme = null;
let notificados = new Set();
let iniciado = false;

/** Cria (ou destrava) o áudio. Precisa ser chamado a partir de um clique. */
export function desbloquearAudio() {
  try {
    audio ??= new (window.AudioContext || window.webkitAudioContext)();
    if (audio.state === 'suspended') audio.resume();
  } catch {
    audio = null;
  }
}

function bip(freq, inicio, dur) {
  const o = audio.createOscillator();
  const g = audio.createGain();
  o.type = 'square';
  o.frequency.value = freq;
  g.gain.setValueAtTime(0.0001, inicio);
  g.gain.exponentialRampToValueAtTime(0.12, inicio + 0.02);
  g.gain.exponentialRampToValueAtTime(0.0001, inicio + dur);
  o.connect(g).connect(audio.destination);
  o.start(inicio);
  o.stop(inicio + dur + 0.02);
}

/** Toca uma vez o padrão do alarme (dois tons, três vezes). */
export function tocarAlarme() {
  if (!audio || audio.state !== 'running') return;
  const t = audio.currentTime;
  for (let i = 0; i < 3; i++) {
    bip(880, t + i * 0.5, 0.2);
    bip(660, t + i * 0.5 + 0.22, 0.2);
  }
}

function ligarAlarme() {
  if (alarme) return;
  tocarAlarme();
  alarme = setInterval(() => {
    if (!plantao.mudo) tocarAlarme();
  }, 4000);
}

function desligarAlarme() {
  clearInterval(alarme);
  alarme = null;
}

export function alternarMudo() {
  plantao.mudo = !plantao.mudo;
  try {
    localStorage.setItem(CHAVE_MUDO, plantao.mudo ? '1' : '0');
  } catch {
    /* sem armazenamento local */
  }
}

/** Pede permissão de notificação (a partir de um clique). */
export async function pedirNotificacao() {
  try {
    if ('Notification' in window && Notification.permission === 'default') {
      await Notification.requestPermission();
    }
  } catch {
    /* navegador sem suporte */
  }
}

function notificar(inc) {
  if (notificados.has(inc.id + inc.alertaEm)) return;
  notificados.add(inc.id + inc.alertaEm);
  try {
    if ('Notification' in window && Notification.permission === 'granted') {
      const n = new Notification(`🚨 Incidente #${inc.number}`, {
        body: inc.alertTitle,
        tag: 'iz-plantao',
        requireInteraction: true
      });
      n.onclick = () => {
        window.focus();
        n.close();
      };
    }
  } catch {
    /* notificação bloqueada */
  }
}

let tituloOriginal = '';
function piscarTitulo(ligado) {
  if (typeof document === 'undefined') return;
  tituloOriginal ||= document.title;
  document.title = ligado ? '🚨 Incidente! · Incidente Zero' : tituloOriginal;
}

function aplicar(d) {
  plantao.dados = d;
  plantao.carregado = true;
  plantao.diferenca = d.agora - Date.now();
  const inc = d.incidente;
  if (inc && inc.status === 'alertando') {
    notificar(inc);
    if (!plantao.mudo) ligarAlarme();
    piscarTitulo(true);
  } else {
    desligarAlarme();
    piscarTitulo(false);
  }
}

export async function atualizarPlantao() {
  try {
    aplicar(await api.plantao());
  } catch {
    /* servidor fora do ar: tenta de novo no próximo ciclo */
  }
}

/** Começa a acompanhar o plantão (uma vez, no App). */
export function acompanharPlantao() {
  if (iniciado) return;
  iniciado = true;
  atualizarPlantao();
  setInterval(atualizarPlantao, 4000);
  // Qualquer clique destrava o áudio (regra dos navegadores)
  window.addEventListener('pointerdown', desbloquearAudio, { passive: true });
  window.addEventListener('keydown', desbloquearAudio, { passive: true });
}

export const RITMOS = [
  { id: 'calmo', nome: 'Calmo', desc: 'Próximo incidente 20 a 35 min depois de resolver o anterior' },
  { id: 'normal', nome: 'Normal', desc: 'Próximo incidente 8 a 18 min depois de resolver o anterior' },
  { id: 'intenso', nome: 'Intenso', desc: 'Próximo incidente 3 a 7 min depois de resolver o anterior' }
];

export const DURACOES = [
  { min: 30, nome: '30 min' },
  { min: 60, nome: '1 h' },
  { min: 120, nome: '2 h' },
  { min: 240, nome: '4 h' }
];
