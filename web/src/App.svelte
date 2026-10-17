<script>
  import Central from './routes/Central.svelte';
  import Lab from './routes/Lab.svelte';

  // Rotas por hash: #/  e  #/lab/<id>
  function parse() {
    const h = location.hash.replace(/^#\/?/, '');
    const [page, id] = h.split('/');
    if (page === 'lab' && id) return { page: 'lab', id: decodeURIComponent(id) };
    return { page: 'central' };
  }

  let route = $state(parse());

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
{:else}
  <Central />
{/if}
