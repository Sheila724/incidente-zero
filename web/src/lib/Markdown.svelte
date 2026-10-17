<script>
  import { marked } from 'marked';
  import DOMPurify from 'dompurify';

  let { source = '', inline = false } = $props();

  // Cenários podem vir da comunidade: o HTML gerado é sempre sanitizado.
  const html = $derived(
    DOMPurify.sanitize(
      inline ? marked.parseInline(source || '') : marked.parse(source || '', { gfm: true, breaks: true })
    )
  );
</script>

{#if inline}
  <span class="md md-inline">{@html html}</span>
{:else}
  <div class="md">{@html html}</div>
{/if}

<style>
  .md-inline :global(code) { font-size: 12.5px; }
</style>
