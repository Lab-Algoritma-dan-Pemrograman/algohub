<script lang="ts">
  import '../app.css';

  let { children } = $props();
</script>

<div class="flex min-h-screen flex-col">
  <main class="flex flex-1 flex-col">
    {@render children()}
  </main>
  <footer
    class="border-t border-gray-200 bg-surface-muted py-6 text-center text-sm text-ink-caption"
  >
    © {new Date().getFullYear()} Laboratorium Algoritma &amp; Pemrograman
  </footer>
</div>
