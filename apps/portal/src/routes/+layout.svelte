<script lang="ts">
  import '../app.css';

  let { children } = $props();

  const services = [
    { name: 'SIAKAD', href: 'https://siakad.algohub.web.id' },
    { name: 'e-Learning', href: 'https://elearning.algohub.web.id' },
    { name: 'Praktikum', href: 'https://praktikum.algohub.web.id' },
  ];
</script>

<div class="flex min-h-screen flex-col">
  <main class="flex flex-1 flex-col">
    {@render children()}
  </main>
  <footer class="border-t border-border bg-surface-muted">
    <div class="mx-auto flex max-w-5xl flex-col items-center gap-4 px-4 py-8 sm:px-6 lg:px-8">
      <p class="font-display text-sm font-semibold text-ink-heading">
        Satu ekosistem untuk kemudahan akademik.
      </p>
      <nav class="flex flex-wrap items-center justify-center gap-x-6 gap-y-2" aria-label="Layanan">
        {#each services as service (service.href)}
          <a
            href={service.href}
            class="text-sm text-ink-caption transition-colors hover:text-primary"
          >
            {service.name}
          </a>
        {/each}
      </nav>
      <p class="text-xs text-ink-caption">
        © {new Date().getFullYear()} AlgoHub · Laboratorium Algoritma &amp; Pemrograman
      </p>
    </div>
  </footer>
</div>
