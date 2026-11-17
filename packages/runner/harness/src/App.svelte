<script lang="ts">
  import { onMount } from 'svelte';
  import { Button, CodeEditor, Terminal } from '@algohub/ui';
  import { getRunner, messages, type Language, type RunResult } from '../../src';
  import { samples } from './samples';
  const runner = getRunner({
    compilerBaseUrl: import.meta.env.PUBLIC_COMPILER_BASE_URL ?? '/compiler/v1/',
  });
  let sampleId = $state(samples[0].id);
  let language = $state<Language>(samples[0].language);
  let code = $state(samples[0].code);
  let running = $state(false);
  let terminal: ReturnType<typeof Terminal> | undefined = $state();
  const isolated = globalThis.crossOriginIsolated === true;
  function selectSample(id: string) {
    const sample = samples.find((s) => s.id === id);
    if (!sample) return;
    sampleId = sample.id;
    language = sample.language;
    code = sample.code;
  }
  function selectLanguage(lang: Language) {
    const sample = samples.find((s) => s.language === lang);
    if (sample) selectSample(sample.id);
  }
  function report(result: RunResult) {
    const t = terminal;
    if (!t) return;
    switch (result.status) {
      case 'compile-error':
        t.writeln('❌ \x1b[1;31mCompilation Error:\x1b[0m');
        t.write(`\x1b[31m${result.output || result.error || messages.compileFailed}\x1b[0m\r\n`);
        break;
      case 'ok':
        if (result.images.length)
          t.writeln(`\x1b[2m[${result.images.length} grafik dirender]\x1b[0m`);
        t.writeln(`\r\n\x1b[1;32m✓ ${messages.done}\x1b[0m`);
        break;
      case 'error':
        t.write(`\r\n\x1b[33m⚠ ${result.error}\x1b[0m\r\n`);
        break;
      case 'timeout':
        t.writeln(`\r\n\x1b[1;31m❌ ${messages.timeout}\x1b[0m`);
        break;
      case 'stopped':
        t.writeln(`\r\n\x1b[1;31m⚠ ${messages.stopped}\x1b[0m`);
        break;
      case 'unavailable':
        t.writeln(`\x1b[33m⚠ ${result.error}\x1b[0m`);
        break;
    }
  }
  async function run() {
    if (running) return;
    running = true;
    const lang = language;
    terminal?.clear();
    if (runner.status[lang] !== 'ready')
      terminal?.writeln(
        `\x1b[2m${lang === 'c' ? messages.loadingC : messages.loadingPython}\x1b[0m`,
      );
    terminal?.focus();
    try {
      report(await runner.run(lang, code, { onOutput: (data) => terminal?.write(data) }));
    } finally {
      running = false;
    }
  }
  onMount(() => {
    runner.prefetch();
    terminal?.writeln(`\x1b[2m${messages.idle}\x1b[0m`);
    if (!runner.interactive) terminal?.writeln(`\x1b[33m⚠ ${messages.noInteractive}\x1b[0m`);
  });
</script>

<main class="mx-auto flex max-w-5xl flex-col gap-4 p-6">
  <header class="flex flex-wrap items-center gap-3">
    <h1 class="font-display text-ink-heading mr-auto text-xl font-bold">Harness @algohub/runner</h1>
    <select
      data-testid="language"
      class="border-border h-9 rounded-md border bg-white px-2 text-sm"
      value={language}
      onchange={(e) => selectLanguage(e.currentTarget.value as Language)}
    >
      <option value="c">C</option>
      <option value="python">Python</option>
    </select>
    <select
      data-testid="sample"
      class="border-border h-9 rounded-md border bg-white px-2 text-sm"
      value={sampleId}
      onchange={(e) => selectSample(e.currentTarget.value)}
    >
      {#each samples as sample (sample.id)}
        <option value={sample.id}>{sample.label}</option>
      {/each}
    </select>
    <Button data-testid="run" disabled={running} onclick={run}>Run</Button>
    <Button data-testid="stop" variant="outline" disabled={!running} onclick={() => runner.stop()}>
      Stop
    </Button>
  </header>
  <p class="text-ink-caption text-xs">
    crossOriginIsolated: <span data-testid="isolated">{String(isolated)}</span> · status:
    <span data-testid="status">{runner.status[language]}</span>
  </p>
  <div class="overflow-hidden rounded-xl border border-zinc-800">
    <CodeEditor bind:value={code} {language} onrun={run} />
  </div>
  <Terminal
    bind:this={terminal}
    data-testid="terminal"
    inputEnabled={running}
    oninputline={(line) => runner.sendInput(line)}
  />
</main>
