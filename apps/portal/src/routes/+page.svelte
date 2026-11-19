<script lang="ts">
  import { Book, Building2, Code, ArrowRight } from '@lucide/svelte';

  const apps = [
    {
      num: '01',
      name: 'SIAKAD',
      type: 'Sistem Informasi Akademik',
      year: '2025',
      href: 'https://siakad.algohub.web.id',
      icon: Building2,
      image: '/framer/Q7jMjcix7IzYGoyjSGEfEGOYLc_1.png',
      description:
        'Administrasi akademik mahasiswa dan dosen: KRS, nilai, dan jadwal dalam satu sistem.',
    },
    {
      num: '02',
      name: 'E-Learning',
      type: 'Pembelajaran Daring',
      year: '2025',
      href: 'https://elearning.algohub.web.id',
      icon: Book,
      image: '/framer/dKodz01ydBwO8nWspPrd33iNk.png',
      description: 'Kelas daring, materi, tugas, dan diskusi untuk pembelajaran fleksibel.',
    },
    {
      num: '03',
      name: 'Praktikum',
      type: 'Laboratorium Digital',
      year: '2025',
      href: 'https://praktikum.algohub.web.id',
      icon: Code,
      image: '/framer/wyARw2JXz4syDnnXOZtxsvFzhg_1.png',
      description: 'Modul dan laporan praktikum dengan pendampingan asisten lab.',
    },
  ];
</script>

<svelte:head>
  <title>AlgoHub · Lab Algoritma &amp; Pemrograman</title>
  <meta
    name="description"
    content="Satu ekosistem digital untuk SIAKAD, e-Learning, dan Praktikum — dirancang agar kegiatan akademik terasa ringan, cepat, dan andal."
  />
</svelte:head>

<!-- Hero: dua kolom, teks kiri + panel gambar kanan. Persis mockup Framer. -->
<section class="hero">
  <div class="hero-inner">
    <div class="anim-intro copy">
      <div class="headline">
        <p class="welcome">Welcome</p>
        <h1 class="name">AlgoHub</h1>
        <p class="role">
          One platform for
          <span class="role-accent">SIAKAD, e-Learning &amp; Praktikum</span>
        </p>
      </div>
      <p class="intro">
        A single digital ecosystem uniting SIAKAD, e-Learning, and Praktikum. Designed to make
        academic activities feel effortless, fast, and reliable.
      </p>
      <div class="anim-intro actions" style="--i: 1">
        <a href={apps[0].href} class="btn-fill" target="_blank" rel="noreferrer">
          Masuk SIAKAD <ArrowRight size={16} aria-hidden="true" />
        </a>
        <a href={apps[1].href} class="btn-line" target="_blank" rel="noreferrer">
          Buka e-Learning
        </a>
      </div>
    </div>

    <div class="anim-intro pic-col" style="--i: 2">
      <div class="portrait">
        <img
          src="/framer/HTce1u8lCmCOsP0X7kjn8omRjko_1.png"
          alt="Visual ekosistem AlgoHub"
          loading="lazy"
        />
        <span class="portrait-name"><b>#</b>Artificial_neural_network.py</span>
        <span class="portrait-size">460 × 575</span>
      </div>
    </div>
  </div>
</section>

<!-- Integrated Platform -->
<section class="platform">
  <div class="platform-inner">
    <div class="anim-reveal head">
      <div class="head-left">
        <h2 class="head-title">Integrated Platform</h2>
        <p class="head-sub">Tiga layanan utama AlgoHub dalam satu ekosistem.</p>
      </div>
      <a href={apps[2].href} class="btn-line btn-sm" target="_blank" rel="noreferrer">
        Buka Praktikum
      </a>
    </div>

    <ul class="grid">
      {#each apps as app, i (app.href)}
        <li class="anim-reveal" style="--i: {i}">
          <a href={app.href} class="card" target="_blank" rel="noreferrer">
            <div class="card-cover">
              <img src={app.image} alt={app.name} loading="lazy" />
            </div>
            <div class="card-body">
              <div class="card-meta">
                <h3 class="card-title">{app.name}</h3>
                <p class="card-info">{app.type} · {app.year}</p>
              </div>
              <p class="card-summary">{app.description}</p>
            </div>
          </a>
        </li>
      {/each}
    </ul>
  </div>
</section>

<style>
  /* Ground/Page dari token Framer */
  body {
    background-color: #f4f4f4;
    color: #0a0a0a;
    font-family: 'Geist', 'Inter', ui-sans-serif, system-ui, sans-serif;
  }

  /* ---------- Hero ---------- */
  .hero {
    background-color: #f4f4f4;
    padding: 120px 72px 60px;
  }
  .hero-inner {
    max-width: 1776px;
    min-height: 100vh;
    display: flex;
    flex-flow: row;
    align-items: center;
    gap: 60px;
    margin: 0 auto;
  }
  .copy {
    flex: 1 0 0;
    display: flex;
    flex-flow: column;
    align-items: flex-start;
    gap: 18px;
  }
  .headline {
    display: flex;
    flex-flow: column;
    gap: 0;
    padding: 2px 0 0;
  }
  .welcome {
    color: #8a8a8a;
    font-size: 36px;
    font-weight: 600;
    letter-spacing: -0.045em;
    line-height: 1.08em;
  }
  .name {
    color: #0a0a0a;
    font-size: 90px;
    font-weight: 600;
    letter-spacing: -0.045em;
    line-height: 1.08em;
  }
  .role {
    color: #0a0a0a;
    font-size: 36px;
    font-weight: 600;
    letter-spacing: -0.045em;
    line-height: 1.08em;
    padding-top: 4px;
  }
  .role-accent {
    color: #8a1538;
  }
  .intro {
    color: #605f5f;
    font-size: 16px;
    font-weight: 400;
    line-height: 1.55em;
    max-width: 480px;
  }
  .actions {
    display: flex;
    flex-wrap: wrap;
    gap: 12px;
    padding: 8px 0 0;
  }

  /* ---------- Tombol (pill) ---------- */
  .btn-fill,
  .btn-line {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    padding: 14px 26px;
    border-radius: 999px;
    font-size: 14px;
    font-weight: 500;
    letter-spacing: -0.01em;
    line-height: 1.5em;
    text-decoration: none;
    transition:
      transform 0.2s ease,
      opacity 0.2s ease;
  }
  .btn-fill {
    background-color: #0a0a0a;
    color: #f4f4f4;
  }
  .btn-line {
    border: 1px solid rgba(10, 10, 10, 0.1);
    background-color: transparent;
    color: #0a0a0a;
  }
  .btn-fill:hover,
  .btn-line:hover {
    transform: translateY(-1px);
    opacity: 0.9;
  }
  .btn-sm {
    padding: 11px 20px;
  }

  /* ---------- Panel gambar kanan ---------- */
  .pic-col {
    display: flex;
    flex-flow: column;
    align-items: flex-end;
    width: 45%;
  }
  .portrait {
    position: relative;
    aspect-ratio: 0.8;
    width: 100%;
    max-width: 460px;
    background-color: #ececec;
    border-radius: 20px;
    overflow: clip;
  }
  .portrait img {
    position: absolute;
    inset: 0;
    width: 100%;
    height: 100%;
    object-fit: cover;
    object-position: 50% 30%;
  }
  .portrait-name {
    position: absolute;
    left: 12px;
    bottom: 12px;
    z-index: 2;
    background: rgba(255, 255, 255, 0.9);
    color: #0a0a0a;
    border-radius: 6px;
    padding: 4px 8px;
    font-family: 'Inter', sans-serif;
    font-size: 12px;
    font-weight: 400;
    line-height: 1.35em;
  }
  .portrait-size {
    position: absolute;
    right: 12px;
    top: 12px;
    z-index: 2;
    background: #8a1538;
    color: #0a0a0a;
    border-radius: 5px;
    padding: 3px 8px;
    font-family: 'Inter', sans-serif;
    font-size: 12px;
    font-weight: 600;
    line-height: 1.35em;
  }

  /* ---------- Integrated Platform ---------- */
  .platform {
    padding: 108px 72px;
  }
  .platform-inner {
    max-width: 1776px;
    display: flex;
    flex-flow: column;
    gap: 52px;
    margin: 0 auto;
  }
  .head {
    display: flex;
    flex-wrap: wrap;
    align-items: flex-end;
    justify-content: space-between;
    gap: 24px;
  }
  .head-left {
    display: flex;
    flex-flow: column;
    gap: 14px;
  }
  .head-title {
    color: #0a0a0a;
    font-size: 56px;
    font-weight: 600;
    letter-spacing: -0.045em;
    line-height: 1.08em;
  }
  .head-sub {
    color: #605f5f;
    font-size: 15px;
    font-weight: 400;
    line-height: 1.55em;
    max-width: 544px;
  }

  /* ---------- Kartu ---------- */
  .grid {
    display: grid;
    grid-template-columns: repeat(3, minmax(0, 1fr));
    gap: 16px;
    width: 100%;
    list-style: none;
    padding: 0;
    margin: 0;
  }
  .card {
    display: flex;
    flex-flow: column;
    gap: 16px;
    width: 100%;
    text-decoration: none;
    color: inherit;
  }
  .card-cover {
    aspect-ratio: 1.33333;
    width: 100%;
    overflow: clip;
    border-radius: 20px;
    background-color: #ececec;
  }
  .card-cover img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    object-position: center;
    transition: transform 0.4s cubic-bezier(0.16, 1, 0.3, 1);
  }
  .card:hover .card-cover img {
    transform: scale(1.04);
  }
  .card-body {
    display: flex;
    flex-flow: column;
    gap: 4px;
    width: 100%;
  }
  .card-meta {
    display: flex;
    flex-flow: row;
    align-items: flex-end;
    justify-content: space-between;
    gap: 16px;
    width: 100%;
  }
  .card-title {
    color: #0a0a0a;
    font-size: 18px;
    font-weight: 500;
    letter-spacing: -0.02em;
    line-height: 1.35em;
    margin: 0;
  }
  .card-info {
    color: #8a8a8a;
    font-size: 12px;
    font-weight: 400;
    line-height: 1.3em;
    margin: 0;
    white-space: nowrap;
  }
  .card-summary {
    color: #605f5f;
    font-size: 14px;
    font-weight: 400;
    line-height: 1.5em;
    margin: 0;
  }

  /* ---------- Responsive ---------- */
  @media (max-width: 1199.98px) {
    .hero,
    .platform {
      padding-left: 48px;
      padding-right: 48px;
    }
    .name {
      font-size: 65px;
    }
    .welcome,
    .role {
      font-size: 28px;
    }
    .head-title {
      font-size: 43px;
    }
  }
  @media (max-width: 809.98px) {
    .hero {
      padding: 96px 24px 48px;
      min-height: auto;
    }
    .hero-inner {
      flex-flow: column;
      align-items: flex-start;
      gap: 40px;
    }
    .copy,
    .pic-col {
      width: 100%;
    }
    .name {
      font-size: 36px;
    }
    .welcome,
    .role {
      font-size: 20px;
    }
    .platform {
      padding: 72px 24px;
    }
    .head-title {
      font-size: 34px;
    }
    .grid {
      grid-template-columns: 1fr;
    }
  }
</style>
