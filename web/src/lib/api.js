// Cliente da API do `iz ui`.

async function request(method, path, data) {
  const res = await fetch(`/api${path}`, {
    method,
    headers: { Accept: 'application/json', ...(data ? { 'Content-Type': 'application/json' } : {}) },
    ...(data ? { body: JSON.stringify(data) } : {})
  });
  let body = null;
  try {
    body = await res.json();
  } catch {
    /* resposta sem JSON */
  }
  if (!res.ok) {
    const err = new Error(body?.error || `HTTP ${res.status}`);
    err.status = res.status;
    err.body = body;
    throw err;
  }
  return body;
}

export const api = {
  status: () => request('GET', '/status'),
  scenarios: () => request('GET', '/scenarios'),
  scenario: (id) => request('GET', `/scenarios/${encodeURIComponent(id)}`),
  start: (id) => request('POST', `/scenarios/${encodeURIComponent(id)}/start`),
  stop: (id) => request('POST', `/scenarios/${encodeURIComponent(id)}/stop`),
  check: (id) => request('POST', `/scenarios/${encodeURIComponent(id)}/check`),
  steps: (id) => request('GET', `/scenarios/${encodeURIComponent(id)}/steps`),
  metrics: (id) => request('GET', `/scenarios/${encodeURIComponent(id)}/metrics`),
  hint: (id, n) => request('POST', `/scenarios/${encodeURIComponent(id)}/hints/${n}`),
  solution: (id) => request('GET', `/scenarios/${encodeURIComponent(id)}/solution`),
  postmortem: (id, tz = '') => request('GET', `/scenarios/${encodeURIComponent(id)}/postmortem?tz=${encodeURIComponent(tz)}`),
  plantao: () => request('GET', '/plantao'),
  assumirPlantao: (dados) => request('POST', '/plantao', dados),
  encerrarPlantao: () => request('POST', '/plantao/encerrar'),
  reconhecer: () => request('POST', '/plantao/reconhecer'),
  savePostmortem: (id, pm, tz = '') =>
    request('PUT', `/scenarios/${encodeURIComponent(id)}/postmortem?tz=${encodeURIComponent(tz)}`, pm)
};

export function postmortemURL(id, tz = '') {
  return `/api/scenarios/${encodeURIComponent(id)}/postmortem.md?tz=${encodeURIComponent(tz)}`;
}

export function terminalURL(id) {
  const proto = location.protocol === 'https:' ? 'wss' : 'ws';
  return `${proto}://${location.host}/api/scenarios/${encodeURIComponent(id)}/terminal`;
}
