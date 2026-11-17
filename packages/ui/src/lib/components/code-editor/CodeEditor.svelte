<script lang="ts">
  import { onDestroy, onMount } from 'svelte';
  import type { Monaco } from '@monaco-editor/loader';

  let {
    value = $bindable(''),
    language = 'c',
    readonly = false,
    height = '350px',
    onrun,
  }: {
    value?: string;
    language?: 'c' | 'python';
    readonly?: boolean;
    height?: string;
    onrun?: () => void;
  } = $props();

  let el: HTMLDivElement | undefined = $state();
  let monacoRef: Monaco | null = null;
  let editor: ReturnType<Monaco['editor']['create']> | null = $state(null);
  let loadError = $state('');

  onMount(async () => {
    const loader = (await import('@monaco-editor/loader')).default;
    let monaco: Monaco;
    try {
      monaco = await loader.init();
    } catch {
      loadError =
        'Editor kode gagal dimuat (CDN terblokir di jaringan/ISP Anda). Coba muat ulang halaman atau gunakan koneksi lain.';
      return;
    }
    if (!monaco?.editor) {
      loadError = 'Editor kode gagal dimuat. Coba muat ulang halaman.';
      return;
    }
    monacoRef = monaco;
    if (!el) return;
    try {
      editor = monaco.editor.create(el, {
        value,
        language,
        readOnly: readonly,
        automaticLayout: true,
        minimap: { enabled: false },
        fontSize: 14,
        scrollBeyondLastLine: false,
        theme: 'vs-dark',
      });
    } catch {
      loadError = 'Editor kode gagal dibuat. Coba muat ulang halaman.';
      return;
    }
    editor.onDidChangeModelContent(() => {
      value = editor?.getValue() ?? '';
    });
    editor.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyCode.Enter, () => onrun?.());
  });

  $effect(() => {
    // `value` wajib dibaca tanpa syarat di baris pertama: kalau dibaca setelah guard
    // `editor &&` (false saat init), Svelte tidak mencatatnya sebagai dependency dan
    // effect ini tidak akan jalan lagi saat halaman induk mengganti isi editor.
    const v = value;
    if (editor && v !== editor.getValue()) editor.setValue(v ?? '');
  });

  $effect(() => {
    const ro = readonly;
    editor?.updateOptions({ readOnly: ro });
  });

  $effect(() => {
    const lang = language;
    const model = editor?.getModel();
    if (monacoRef && model) monacoRef.editor.setModelLanguage(model, lang);
  });

  onDestroy(() => {
    editor?.dispose();
  });
</script>

{#if loadError}
  <div class="p-6 text-center" style="height: {height};">
    <p class="text-sm font-bold text-red-400">{loadError}</p>
    <button
      type="button"
      class="mt-3 h-8 rounded-lg border border-zinc-700 bg-zinc-800 px-4 py-1 text-xs font-bold text-zinc-200 hover:bg-zinc-700"
      onclick={() => location.reload()}
    >
      Muat Ulang
    </button>
  </div>
{:else}
  <div bind:this={el} style="height: {height};" class="overflow-hidden bg-[#1e1e1e]"></div>
{/if}
