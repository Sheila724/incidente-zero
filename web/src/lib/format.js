export const SEVERITY = {
  critica: { label: 'Crítica', pill: 'CRÍTICA', color: 'var(--red)', ink: '#1C0507' },
  alta: { label: 'Alta', pill: 'ALTA', color: 'var(--violet)', ink: '#1B1036' },
  media: { label: 'Média', pill: 'MÉDIA', color: 'var(--blue)', ink: '#071A33' },
  baixa: { label: 'Baixa', pill: 'BAIXA', color: 'var(--muted)', ink: '#0A0E13' }
};

export const sev = (s) => SEVERITY[s] || SEVERITY.media;

export const LEVELS = [
  { id: 'iniciante', name: 'Iniciante', title: 'Trilha iniciante', desc: 'Linux e troubleshooting', planned: 8, color: 'var(--accent)' },
  { id: 'intermediario', name: 'Intermediário', title: 'Trilha intermediária', desc: 'Rede, web, banco e Zabbix', planned: 7, color: 'var(--blue)' },
  { id: 'avancado', name: 'Avançado', title: 'Trilha avançada', desc: 'Kubernetes e incidentes em cascata', planned: 10, color: 'var(--red)' }
];

export function duration(totalSeconds) {
  const s = Math.max(0, Math.floor(totalSeconds || 0));
  const h = Math.floor(s / 3600);
  const m = Math.floor((s % 3600) / 60);
  const r = s % 60;
  if (h > 0) return `${h}h ${String(m).padStart(2, '0')}m`;
  return `${m}m ${String(r).padStart(2, '0')}s`;
}

export function clock(totalSeconds) {
  const s = Math.max(0, Math.floor(totalSeconds || 0));
  const h = Math.floor(s / 3600);
  const m = Math.floor((s % 3600) / 60);
  const r = s % 60;
  const mm = String(m).padStart(2, '0');
  const ss = String(r).padStart(2, '0');
  return h > 0 ? `${h}:${mm}:${ss}` : `${mm}:${ss}`;
}

export function today() {
  return new Intl.DateTimeFormat('pt-BR', { weekday: 'short', day: '2-digit', month: 'short' })
    .format(new Date())
    .replace(/\./g, '');
}
