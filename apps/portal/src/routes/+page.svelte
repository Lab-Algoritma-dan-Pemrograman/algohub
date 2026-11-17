<script lang="ts">
  import { Book, Building2, Code } from '@lucide/svelte';
  import { Card, CardDescription, CardHeader, CardTitle } from '@algohub/ui';

  const apps = [
    { name: 'SIAKAD', href: 'https://siakad.algohub.web.id', icon: Building2 },
    { name: 'E-Learning', href: 'https://elearning.algohub.web.id', icon: Book },
    { name: 'Praktikum', href: 'https://praktikum.algohub.web.id', icon: Code },
  ];
</script>

<section
  class="relative flex flex-1 items-center justify-center overflow-hidden bg-[url('/bg_login.jpg')] bg-cover bg-center px-4 py-16"
>
  <div class="absolute inset-0 bg-slate-950/65"></div>
  <div class="relative z-10 flex w-full max-w-4xl flex-col items-center gap-10 text-center">
    <img src="/logo_new.png" alt="Logo Lab AP" class="h-14 w-auto rounded-xl bg-white p-2" />
    <div class="space-y-4">
      <h1 class="text-4xl font-extrabold tracking-tight text-white md:text-5xl">
        Laboratorium Algoritma &amp; Pemrograman
      </h1>
      <p class="mx-auto max-w-2xl text-sm font-semibold leading-relaxed text-slate-200">
        Laboratorium Algoritma &amp; Pemrograman merupakan pusat riset, praktikum, dan pengembangan
        skill komputasional bagi mahasiswa. Kami menyediakan sarana terpadu untuk menunjang kegiatan
        akademis dan kreativitas mahasiswa di bidang pemrograman.
      </p>
    </div>
    <ul class="grid w-full grid-cols-1 gap-6 sm:grid-cols-3">
      {#each apps as app (app.href)}
        <li>
          <a
            href={app.href}
            class="group block rounded-xl focus-visible:ring-4 focus-visible:ring-white/60 focus-visible:outline-none"
          >
            <Card class="h-full transition-all group-hover:-translate-y-0.5 group-hover:shadow-md">
              <CardHeader class="items-center justify-items-center gap-3 text-center">
                <span
                  class="flex size-12 items-center justify-center rounded-xl bg-primary text-primary-foreground"
                >
                  <app.icon size={24} aria-hidden="true" />
                </span>
                <CardTitle class="text-lg text-ink-heading">{app.name}</CardTitle>
                <CardDescription>{new URL(app.href).host}</CardDescription>
              </CardHeader>
            </Card>
          </a>
        </li>
      {/each}
    </ul>
  </div>
</section>
