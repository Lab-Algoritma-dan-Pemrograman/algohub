<script lang="ts">
  import { Book, Building2, Code, ArrowRight } from '@lucide/svelte';
  import { Badge, Button, Card, CardDescription, CardHeader, CardTitle } from '@algohub/ui';
  import TypeRotator from '../lib/components/TypeRotator.svelte';

  const apps = [
    {
      name: 'SIAKAD',
      href: 'https://siakad.algohub.web.id',
      icon: Building2,
      description:
        'Administrasi akademik mahasiswa dan dosen: KRS, nilai, dan jadwal dalam satu sistem.',
    },
    {
      name: 'E-Learning',
      href: 'https://elearning.algohub.web.id',
      icon: Book,
      description: 'Kelas daring, materi, tugas, dan diskusi untuk pembelajaran fleksibel.',
    },
    {
      name: 'Praktikum',
      href: 'https://praktikum.algohub.web.id',
      icon: Code,
      description: 'Modul dan laporan praktikum terstruktur dengan pendampingan asisten lab.',
    },
  ];
</script>

<svelte:head>
  <title>Lab Algoritma &amp; Pemrograman</title>
  <meta
    name="description"
    content="Satu ekosistem digital untuk kemudahan akademik: SIAKAD, e-Learning, dan Praktikum."
  />
</svelte:head>

<!-- Hero -->
<section class="px-4 pt-24 pb-16 sm:px-6 lg:px-8">
  <div class="mx-auto flex max-w-5xl flex-col items-center gap-8 text-center">
    <img
      src="/logo_new.png"
      alt="Logo Lab AP"
      class="h-16 w-auto rounded-xl bg-white p-2 shadow-xs ring-1 ring-border"
    />
    <div class="space-y-5">
      <Badge variant="outline" class="text-ink-caption">Ekosistem Digital Kampus</Badge>
      <h1 class="font-display text-4xl font-extrabold tracking-tight text-ink-heading md:text-5xl">
        Selamat datang di <span class="text-primary">AlgoHub</span>
      </h1>
      <p class="font-display text-2xl font-semibold tracking-tight text-ink-heading md:text-3xl">
        One platform for
        <TypeRotator words={['SIAKAD', 'e-Learning', 'Praktikum', 'Portal']} />
      </p>
      <p class="mx-auto max-w-2xl text-base leading-relaxed text-ink-body">
        Laboratorium Algoritma &amp; Pemrograman merupakan pusat riset, praktikum, dan pengembangan
        skill komputasional bagi mahasiswa. Satu ekosistem digital yang menyatukan SIAKAD,
        e-Learning, dan Praktikum agar kegiatan akademik terasa ringan, cepat, dan mudah dipercaya.
      </p>
    </div>
    <div class="flex flex-wrap items-center justify-center gap-3">
      <Button href={apps[0].href} size="lg">
        Masuk SIAKAD <ArrowRight aria-hidden="true" />
      </Button>
      <Button href={apps[1].href} variant="outline" size="lg">Buka e-Learning</Button>
    </div>
  </div>
</section>

<!-- Platform -->
<section class="px-4 pb-28 sm:px-6 lg:px-8">
  <div class="mx-auto flex max-w-5xl flex-col items-center gap-10">
    <div class="space-y-2 text-center">
      <h2 class="font-display text-2xl font-bold tracking-tight text-ink-heading md:text-3xl">
        Integrated Platform
      </h2>
      <p class="text-sm text-ink-caption">Tiga layanan utama AlgoHub dalam satu ekosistem.</p>
    </div>
    <ul class="grid w-full grid-cols-1 gap-6 sm:grid-cols-3">
      {#each apps as app, i (app.href)}
        <li>
          <a
            href={app.href}
            class="group block rounded-xl focus-visible:ring-4 focus-visible:ring-ring/40 focus-visible:outline-none"
          >
            <Card class="h-full transition-all group-hover:-translate-y-0.5 group-hover:shadow-md">
              <CardHeader class="items-center gap-3 text-center">
                <span
                  class="flex size-12 items-center justify-center rounded-xl bg-primary text-primary-foreground"
                >
                  <app.icon size={24} aria-hidden="true" />
                </span>
                <CardTitle class="text-lg text-ink-heading">
                  <span class="text-ink-caption">0{i + 1}</span>
                  {app.name}
                </CardTitle>
                <CardDescription>{app.description}</CardDescription>
              </CardHeader>
            </Card>
          </a>
        </li>
      {/each}
    </ul>
  </div>
</section>
