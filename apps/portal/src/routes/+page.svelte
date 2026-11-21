<script lang="ts">
  import { Book, Building2, FlaskConical } from '@lucide/svelte';

  const apps = [
    {
      name: 'SIAKAD',
      type: 'Sistem Informasi Akademik',
      year: '2025',
      href: 'https://siakad.algohub.web.id',
      icon: Building2,
      image: '/framer/Q7jMjcix7IzYGoyjSGEfEGOYLc_1.png',
      summary: 'Administrasi akademik mahasiswa dan dosen dalam satu tempat.',
    },
    {
      name: 'e-Learning',
      type: 'Pembelajaran Online',
      year: '2025',
      href: 'https://elearning.algohub.web.id',
      icon: Book,
      image: '/framer/dKodz01ydBwO8nWspPrd33iNk.png',
      summary: 'Kelas, materi, dan tugas daring yang terstruktur.',
    },
    {
      name: 'Praktikum',
      type: 'Laboratorium Digital',
      year: '2025',
      href: 'https://praktikum.algohub.web.id',
      icon: FlaskConical,
      image: '/framer/wyARw2JXz4syDnnXOZtxsvFzhg_1.png',
      summary: 'Modul dan laporan praktikum dengan pendampingan asisten.',
    },
  ];

  // 8 resize handles: 4 corners + 4 edges.
  const handles = [
    { pos: 'nw', cursor: 'nwse-resize' },
    { pos: 'n', cursor: 'ns-resize' },
    { pos: 'ne', cursor: 'nesw-resize' },
    { pos: 'e', cursor: 'ew-resize' },
    { pos: 'se', cursor: 'nwse-resize' },
    { pos: 's', cursor: 'ns-resize' },
    { pos: 'sw', cursor: 'nesw-resize' },
    { pos: 'w', cursor: 'ew-resize' },
  ];

  const headlineLines = [
    { text: 'Welcome', color: '#8a8a8a', perLetter: false },
    { text: 'AlgoHub', color: '#0a0a0a', perLetter: true },
    { text: 'One platform for', color: '#0a0a0a', perLetter: false },
  ];
</script>

<svelte:head>
  <title>AlgoHub · Lab Algoritma &amp; Pemrograman</title>
  <meta
    name="description"
    content="Satu ekosistem digital untuk SIAKAD, e-Learning, dan Praktikum — dirancang agar kegiatan akademik terasa ringan, cepat, dan andal."
  />
</svelte:head>

<!-- Hero -->
<section class="hero">
  <div class="hero-inner">
    <div class="copy">
      <div class="headline">
        {#each headlineLines as line, li (li)}
          <p class="hl" style="color: {line.color}; --d: {li * 0.08}s">
            {#if line.perLetter}
              {#each line.text.split('') as ch, ci (ci)}<span
                  class="ltr"
                  style="--d: {li * 0.08 + ci * 0.03}s">{ch}</span
                >{/each}
            {:else}
              {line.text}
            {/if}
          </p>
        {/each}
        <p class="hl accent" style="--d: 0.24s">
          <span class="accent-word">SIAKAD</span><span class="caret" aria-hidden="true"></span>
        </p>
      </div>
      <p class="intro">
        A single digital ecosystem uniting SIAKAD, e-Learning, and Practicum. Designed to make
        academic activities feel effortless, fast, and reliable
      </p>
      <div class="actions">
        <a href={apps[0].href} class="btn-fill" target="_blank" rel="noreferrer">Masuk SIAKAD</a>
        <a href={apps[1].href} class="btn-line" target="_blank" rel="noreferrer">Buka e-Learning</a>
      </div>
    </div>

    <div class="pic-col">
      <div class="portrait">
        <img
          src="/framer/HTce1u8lCmCOsP0X7kjn8omRjko_1.png"
          alt="Visual ekosistem AlgoHub"
          loading="lazy"
        />
        <span class="frame-border" aria-hidden="true"></span>
        <span class="portrait-name"><b>#</b>Artificial_neural_network.py</span>
        {#each handles as h (h.pos)}
          <span class="handle {h.pos}" style="cursor: {h.cursor}" aria-hidden="true"></span>
        {/each}
        <span class="portrait-hint">Drag a handle to resize me</span>
        <span class="portrait-size">460 × 575</span>
      </div>
      <div class="cursor-chip" aria-hidden="true">
        <svg
          class="cursor-arrow"
          width="14"
          height="14"
          viewBox="0 0 24 24"
          fill="rgb(10,10,10)"
          stroke="#fff"
          stroke-width="1.2"
          stroke-linejoin="round"
        >
          <path d="M4 3l7 17 2.5-6.5L20 11 4 3z" />
        </svg>
        King Nopal
      </div>
    </div>
  </div>
</section>

<!-- Integrated Platform -->
<section class="platform">
  <div class="platform-inner">
    <div class="head">
      <div class="head-left">
        <h2 class="head-title">
          {#each ['Integrated', 'Platform'] as word, i (i)}<span class="w" style="--d: {i * 0.08}s"
              >{word}</span
            >{/each}
        </h2>
        <p class="head-sub">Tiga layanan utama Algohub dalam satu ekosistem.</p>
      </div>
      <a href={apps[2].href} class="btn-line" target="_blank" rel="noreferrer">Buka Praktikum</a>
    </div>

    <ul class="grid">
      {#each apps as app, i (app.href)}
        <li style="--d: {i * 0.08}s">
          <a href={app.href} class="card" target="_blank" rel="noreferrer">
            <div class="card-cover">
              <img src={app.image} alt={app.name} loading="lazy" />
            </div>
            <div class="card-body">
              <div class="card-meta">
                <h3 class="card-title">{app.name}</h3>
                <p class="card-info">{app.type} · {app.year}</p>
              </div>
              <p class="card-summary">{app.summary}</p>
            </div>
          </a>
        </li>
      {/each}
    </ul>
  </div>
</section>

<style>
  /* ---------- Hero ---------- */
  .hero {
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
    gap: 18px;
  }
  .headline {
    display: flex;
    flex-flow: column;
    gap: 0;
    padding: 2px 0 0;
    margin: 0;
  }
  .hl {
    font-family: 'Geist', sans-serif;
    font-size: 78px;
    font-weight: 600;
    letter-spacing: -0.045em;
    line-height: 1.08em;
    margin: 0;
  }
  .accent {
    color: rgb(30, 41, 59);
  }
  .accent-word {
    white-space: nowrap;
  }
  .caret {
    display: inline-block;
    width: max(2px, 0.04em);
    height: 0.82em;
    margin-left: 0.04em;
    vertical-align: -0.04em;
    background-color: rgb(138, 21, 56);
  }
  @media (prefers-reduced-motion: no-preference) {
    .caret {
      animation: blink 1s step-end infinite;
    }
    .hl {
      animation: fade-up 0.9s cubic-bezier(0.16, 1, 0.3, 1) both;
      animation-delay: var(--d, 0s);
    }
    .ltr {
      display: inline-block;
      animation: letter-reveal 0.9s cubic-bezier(0.16, 1, 0.3, 1) both;
      animation-delay: var(--d, 0s);
    }
    .intro {
      animation: fade-up 0.9s cubic-bezier(0.16, 1, 0.3, 1) both;
      animation-delay: 0.16s;
    }
    .actions {
      animation: fade-up 0.9s cubic-bezier(0.16, 1, 0.3, 1) both;
      animation-delay: 0.24s;
    }
    .pic-col {
      animation: fade-up 0.9s cubic-bezier(0.16, 1, 0.3, 1) both;
      animation-delay: 0.08s;
    }
  }
  @keyframes blink {
    0%,
    100% {
      opacity: 1;
    }
    50% {
      opacity: 0.2;
    }
  }
  @keyframes letter-reveal {
    from {
      transform: translateY(30px);
      opacity: 0.001;
    }
    to {
      transform: translateY(0);
      opacity: 1;
    }
  }
  .intro {
    color: #605f5f;
    font-family: 'Geist', sans-serif;
    font-size: 16px;
    font-weight: 400;
    line-height: 1.55em;
    letter-spacing: 0;
    max-width: 480px;
    margin: 0;
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
    font-family: 'Geist', sans-serif;
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

  /* ---------- Panel gambar kanan ---------- */
  .pic-col {
    position: relative;
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
  }
  .portrait img {
    position: absolute;
    inset: 0;
    width: 100%;
    height: 100%;
    object-fit: cover;
    object-position: 50% 30%;
    border-radius: 20px;
  }
  .frame-border {
    position: absolute;
    inset: 0;
    border: 1.5px solid rgb(10, 10, 10);
    border-radius: 20px;
    pointer-events: none;
    z-index: 3;
  }
  .portrait-name {
    position: absolute;
    left: -1px;
    bottom: calc(100% + 8px);
    z-index: 2;
    white-space: nowrap;
    color: rgb(10, 10, 10);
    font-family: 'Inter', sans-serif;
    font-size: 12px;
    font-weight: 400;
    line-height: 1.35em;
  }
  .portrait-name b {
    color: rgb(138, 138, 138);
    font-weight: 400;
  }
  .portrait-hint {
    position: absolute;
    right: 0;
    top: calc(100% + 14px);
    z-index: 2;
    color: rgb(138, 138, 138);
    font-family: 'Inter', sans-serif;
    font-size: 12px;
    font-weight: 400;
    line-height: 1.35em;
    white-space: nowrap;
  }
  .portrait-size {
    position: absolute;
    left: 50%;
    top: calc(100% + 12px);
    translate: -50% 0;
    z-index: 2;
    background: #d6ff3f;
    color: rgb(10, 10, 10);
    padding: 3px 8px;
    border-radius: 5px;
    font-family: 'Inter', sans-serif;
    font-size: 12px;
    font-weight: 600;
    line-height: 1.35em;
    white-space: nowrap;
  }
  .handle {
    position: absolute;
    width: 12px;
    height: 12px;
    background: #fff;
    border: 1.5px solid rgb(10, 10, 10);
    z-index: 4;
    transition:
      transform 0.15s ease,
      background-color 0.15s ease;
  }
  .handle:hover {
    transform: scale(1.35);
    background-color: #d6ff3f;
  }
  .handle.nw {
    left: -7px;
    top: -7px;
  }
  .handle.n {
    left: calc(50% - 6px);
    top: -7px;
  }
  .handle.ne {
    right: -7px;
    top: -7px;
  }
  .handle.e {
    right: -7px;
    top: calc(50% - 6px);
  }
  .handle.se {
    right: -7px;
    bottom: -7px;
  }
  .handle.s {
    left: calc(50% - 6px);
    bottom: -7px;
  }
  .handle.sw {
    left: -7px;
    bottom: -7px;
  }
  .handle.w {
    left: -7px;
    top: calc(50% - 6px);
  }
  .cursor-chip {
    position: absolute;
    right: -12px;
    bottom: -36px;
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 4px 9px;
    background-color: rgb(10, 10, 10);
    border-radius: 7px 7px 7px 2px;
    color: #fff;
    font-family: 'Inter', sans-serif;
    font-size: 12px;
    font-weight: 600;
    line-height: 1.2em;
    white-space: nowrap;
    z-index: 5;
  }
  .cursor-arrow {
    flex: 0 0 auto;
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
    font-family: 'Geist', sans-serif;
    font-size: 52px;
    font-weight: 600;
    letter-spacing: -0.045em;
    line-height: 1.08em;
    margin: 0;
  }
  .head-title .w {
    display: inline-block;
  }
  .head-sub {
    color: #605f5f;
    font-family: 'Geist', sans-serif;
    font-size: 15px;
    font-weight: 400;
    line-height: 1.55em;
    max-width: 544px;
    margin: 0;
  }

  @media (prefers-reduced-motion: no-preference) {
    .head-title .w {
      animation: letter-reveal 0.9s cubic-bezier(0.16, 1, 0.3, 1) both;
      animation-delay: var(--d, 0s);
    }
    .head-left,
    .head > .btn-line {
      animation: fade-up 0.9s cubic-bezier(0.16, 1, 0.3, 1) both;
      animation-delay: 0.08s;
    }
    .grid li {
      animation: fade-up 0.9s cubic-bezier(0.16, 1, 0.3, 1) both;
      animation-delay: var(--d, 0s);
    }
  }

  /* ---------- Kartu ---------- */
  .grid {
    display: grid;
    grid-template-columns: repeat(3, minmax(50px, 1fr));
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
    transition:
      transform 0.4s cubic-bezier(0.16, 1, 0.3, 1),
      opacity 0.4s ease;
  }
  .card:hover {
    transform: translateY(-4px);
    opacity: 0.92;
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
    font-family: 'Geist', sans-serif;
    font-size: 18px;
    font-weight: 500;
    letter-spacing: -0.02em;
    line-height: 1.35em;
    margin: 0;
  }
  .card-info {
    color: #8a8a8a;
    font-family: 'Geist', sans-serif;
    font-size: 12px;
    font-weight: 400;
    line-height: 1.3em;
    margin: 0;
    white-space: nowrap;
  }
  .card-summary {
    color: #605f5f;
    font-family: 'Geist', sans-serif;
    font-size: 14px;
    font-weight: 400;
    line-height: 1.5em;
    margin: 0;
  }

  /* ---------- Breakpoints ---------- */
  @media (max-width: 1439.98px) {
    .hero {
      gap: 72px;
      padding: 120px 88px 60px;
    }
    .hl {
      font-size: 65px;
    }
    .btn-fill,
    .btn-line {
      padding: 11px 20px;
    }
    .platform {
      padding: 130px 88px;
    }
    .head-title {
      font-size: 43px;
    }
  }
  @media (max-width: 1199.98px) {
    .hero {
      flex-direction: column;
      gap: 40px;
      padding: 110px 48px 60px;
    }
    .hl {
      font-size: 44px;
    }
    .pic-col {
      width: 45%;
    }
    .portrait {
      max-width: 380px;
    }
    .platform {
      padding: 72px 48px;
    }
    .head-title {
      font-size: 34px;
    }
  }
  @media (max-width: 809.98px) {
    .hero {
      padding: 96px 24px 48px;
      min-height: auto;
      flex-flow: column;
      align-items: flex-start;
      gap: 40px;
    }
    .hl {
      font-size: 36px;
    }
    .pic-col {
      width: 100%;
    }
    .btn-fill,
    .btn-line {
      padding: 11px 20px;
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
