<script>
  import { onMount } from 'svelte';
  import TopBar from '../lib/TopBar.svelte';
  import { api } from '../lib/api.js';

  let status = $state(null);
  onMount(() => {
    api.status().then((s) => (status = s)).catch(() => {});
  });

  const passos = [
    { t: 'Escolha um incidente', d: 'Na Central, pegue o próximo recomendado ou qualquer um do catálogo. Cada trilha vai do mais simples ao mais difícil.' },
    { t: 'Leia o alerta e assuma', d: 'O briefing conta o que o monitoramento ou o cliente viu. Ao assumir, um servidor de verdade sobe com a falha já injetada, e o cronômetro começa.' },
    { t: 'Investigue no terminal', d: 'O terminal no navegador está ligado ao servidor do lab. O monitoramento ao vivo (HTTP, porta 80, CPU, memória e disco) muda enquanto você corrige.' },
    { t: 'Verifique', d: 'Clique em Verificar (Ctrl+Enter). A verificação confere a causa raiz, não só o sintoma: se você só escondeu o problema, ela avisa.' },
    { t: 'Escreva o postmortem', d: 'Depois de resolver, registre o que aconteceu. A linha do tempo do seu atendimento já vem pronta, e o texto sai em Markdown.' }
  ];
</script>

<TopBar {status} current="ajuda" />

<main>
  <section class="hero">
    <span class="mono eyebrow">Como funciona</span>
    <h1>Um plantão de verdade, sem o susto de produção</h1>
    <p class="muted">
      O Incidente Zero coloca você diante de servidores quebrados de propósito. Você diagnostica, corrige a causa raiz e
      registra o que aconteceu, como num plantão de DevOps ou SRE.
    </p>
  </section>

  <section class="panel bloco">
    <h2>O ciclo de um incidente</h2>
    <ol class="passos">
      {#each passos as p, i (p.t)}
        <li>
          <span class="num mono">{i + 1}</span>
          <div><strong>{p.t}</strong><p class="muted">{p.d}</p></div>
        </li>
      {/each}
    </ol>
  </section>

  <div class="grade">
    <section class="panel bloco">
      <h2>Modos</h2>
      <dl>
        <dt>Guiado</dt>
        <dd>Passos com perguntas que orientam a investigação. Os passos com verificação automática ficam verdes quando você chega lá.</dd>
        <dt>Plantão</dt>
        <dd>Você assume um turno e os incidentes caem sem aviso, com alarme e notificação. Sem modo guiado. Reconheça em até 5 minutos para não escalar. <a href="#/plantao">Abrir o Plantão</a>.</dd>
      </dl>
    </section>

    <section class="panel bloco">
      <h2>Pontuação</h2>
      <ul>
        <li>Cada incidente vale <strong>100 pontos</strong>.</li>
        <li>Cada dica custa pontos: as primeiras são baratas, a última quase entrega a resposta.</li>
        <li>Resolver dentro do tempo estimado dá <strong>+10 de bônus</strong>.</li>
        <li>Abrir a solução antes de resolver zera a tentativa.</li>
        <li>No plantão, um incidente escalado perde <strong>20 pontos</strong>.</li>
      </ul>
    </section>

    <section class="panel bloco">
      <h2>Avisos da verificação</h2>
      <dl>
        <dt class="v">Isso esconde o sintoma</dt>
        <dd>Você fez algo que tira o alerta, mas cria outro problema: rollback, desligar o serviço, remover uma proteção.</dd>
        <dt class="b">Quase lá</dt>
        <dd>Parte das verificações já passa. O aviso mostra só o que falta.</dd>
      </dl>
    </section>

    <section class="panel bloco">
      <h2>Termos do plantão</h2>
      <dl>
        <dt>MTTA</dt>
        <dd>Tempo médio entre o alerta e o reconhecimento (alguém assumiu).</dd>
        <dt>MTTR</dt>
        <dd>Tempo médio entre o alerta e a resolução.</dd>
        <dt>Escalonamento</dt>
        <dd>Quando ninguém reconhece o alerta a tempo, ele sobe para o próximo nível. Aqui, custa pontos.</dd>
        <dt>Postmortem</dt>
        <dd>O registro do incidente: o que aconteceu, a causa, a evidência e o que fazer para não repetir.</dd>
      </dl>
    </section>

    <section class="panel bloco">
      <h2>Atalhos</h2>
      <ul class="atalhos">
        <li><span class="kbd">Ctrl+Enter</span> verifica o incidente</li>
        <li><span class="kbd">Esc</span> fecha os avisos</li>
        <li><span class="kbd">Ctrl+S</span> salva o postmortem</li>
      </ul>
    </section>

    <section class="panel bloco">
      <h2>Projeto aberto</h2>
      <p class="muted">
        O Incidente Zero é open source. Dá para sugerir cenários, relatar problemas ou contribuir com código no GitHub.
      </p>
      <p><a class="link" href="https://github.com/Sheila724/incidente-zero" target="_blank" rel="noopener noreferrer">github.com/Sheila724/incidente-zero</a></p>
    </section>
  </div>
</main>

<style>
  main { max-width: 1360px; margin: 0 auto; padding: 36px 32px 56px; display: flex; flex-direction: column; gap: 24px; }
  .eyebrow { font-size: 12px; letter-spacing: .12em; color: var(--accent); text-transform: uppercase; }
  .hero h1 { font-size: 32px; line-height: 1.15; margin: 6px 0; letter-spacing: -0.02em; }
  .hero p { margin: 0; max-width: 760px; }
  h2 { font-size: 18px; margin: 0; }
  .bloco { padding: 22px; border-radius: 14px; display: flex; flex-direction: column; gap: 14px; }
  .bloco p { margin: 0; }
  .passos { list-style: none; margin: 0; padding: 0; display: grid; grid-template-columns: repeat(auto-fit, minmax(220px, 1fr)); gap: 16px; }
  .passos li { display: flex; gap: 12px; align-items: flex-start; }
  .passos p { margin: 4px 0 0; font-size: 14px; }
  .num { flex: none; width: 30px; height: 30px; border-radius: 8px; display: grid; place-items: center; background: var(--accent-soft); color: var(--accent); border: 1px solid var(--accent-line); font-weight: 700; }
  .grade { display: grid; grid-template-columns: repeat(auto-fit, minmax(320px, 1fr)); gap: 20px; }
  dl { margin: 0; display: flex; flex-direction: column; gap: 4px; }
  dt { font-weight: 600; margin-top: 8px; }
  dt:first-child { margin-top: 0; }
  dt.v { color: var(--violet); }
  dt.b { color: var(--blue); }
  dd { margin: 0; color: var(--text-2); font-size: 14px; }
  dd a, .link { color: var(--accent); }
  ul { margin: 0; padding-left: 18px; display: flex; flex-direction: column; gap: 6px; color: var(--text-2); font-size: 14px; }
  .atalhos { list-style: none; padding: 0; }
  .atalhos li { display: flex; align-items: center; gap: 10px; }
  @media (max-width: 640px) {
    main { padding: 24px 16px 40px; }
    .hero h1 { font-size: 26px; }
  }
</style>
