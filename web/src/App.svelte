<script>
  import Central from './routes/Central.svelte';
  import Lab from './routes/Lab.svelte';
  import Postmortem from './routes/Postmortem.svelte';
  import Plantao from './routes/Plantao.svelte';
  import Historico from './routes/Historico.svelte';
  import Ajuda from './routes/Ajuda.svelte';
  import AlertaPlantao from './lib/AlertaPlantao.svelte';
  import { acompanharPlantao } from './lib/plantao.svelte.js';

  // Rotas por hash: #/, #/lab/<id>, #/postmortem/<id>, #/plantao, #/historico e #/ajuda
  function parse() {
    const h = location.hash.replace(/^#\/?/, '');
    const [page, id] = h.split('/');
    if (page === 'lab' && id) return { page: 'lab', id: decodeURIComponent(id) };
    if (page === 'postmortem' && id) return { page: 'postmortem', id: decodeURIComponent(id) };
    if (page === 'plantao' || page === 'historico' || page === 'ajuda') return { page };
    return { page: 'central' };
  }

  let route = $state(parse());

  $effect(() => {
    acompanharPlantao();
  });

  $effect(() => {
    const onHash = () => {
      route = parse();
      window.scrollTo(0, 0);
    };
    window.addEventListener('hashchange', onHash);
    return () => window.removeEventListener('hashchange', onHash);
  });
</script>

{#if route.page === 'lab'}
  {#key route.id}
    <Lab id={route.id} />
  {/key}
{:else if route.page === 'postmortem'}
  {#key route.id}
    <Postmortem id={route.id} />
  {/key}
{:else if route.page === 'plantao'}
  <Plantao />
{:else if route.page === 'historico'}
  <Historico />
{:else if route.page === 'ajuda'}
  <Ajuda />
{:else}
  <Central />
{/if}

<AlertaPlantao />
