<script lang="ts">
  // Rotator kata ala komponen RoleEdit di Framer: diketik huruf demi huruf, jeda, hapus, lalu lanjut ke kata berikutnya.
  let {
    words,
    hold = 3,
    interval = 90,
  }: { words: string[]; hold?: number; interval?: number } = $props();

  let index = $state(0);
  let text = $state('');
  let deleting = $state(false);

  $effect(() => {
    const current = words[index % words.length] ?? '';
    let timeout: ReturnType<typeof setTimeout>;

    if (!deleting && text === current) {
      timeout = setTimeout(() => (deleting = true), hold * 1000);
    } else if (deleting && text === '') {
      deleting = false;
      index = (index + 1) % words.length;
      timeout = setTimeout(() => {}, interval);
    } else {
      const next = deleting
        ? current.slice(0, Math.max(0, text.length - 1))
        : current.slice(0, text.length + 1);
      timeout = setTimeout(() => (text = next), interval);
    }

    return () => clearTimeout(timeout);
  });
</script>

<span aria-live="polite">{text}</span><span class="animate-pulse" aria-hidden="true">|</span>
