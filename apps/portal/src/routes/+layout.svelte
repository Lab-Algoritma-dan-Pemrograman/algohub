<script lang="ts">
  import '../app.css';
  import { resolve } from '$app/paths';
  import { Building2, GraduationCap, FlaskConical } from '@lucide/svelte';

  let { children } = $props();

  const services = [
    { name: 'SIAKAD', href: 'https://siakad.algohub.web.id' },
    { name: 'e-Learning', href: 'https://elearning.algohub.web.id' },
    { name: 'Praktikum', href: 'https://praktikum.algohub.web.id' },
  ];

  const dock = [
    { name: 'SIAKAD', key: 'S', href: 'https://siakad.algohub.web.id', icon: Building2 },
    { name: 'e-Learning', key: 'E', href: 'https://elearning.algohub.web.id', icon: GraduationCap },
    { name: 'Praktikum', key: 'P', href: 'https://praktikum.algohub.web.id', icon: FlaskConical },
  ];
</script>

<svelte:head>
  <style>
    @font-face {
      font-family: 'Geist';
      font-style: normal;
      font-weight: 400;
      font-display: swap;
      src: url('/fonts/Geist-latin-400.woff2') format('woff2');
    }
    @font-face {
      font-family: 'Geist';
      font-style: normal;
      font-weight: 500;
      font-display: swap;
      src: url('/fonts/Geist-latin-500.woff2') format('woff2');
    }
    @font-face {
      font-family: 'Geist';
      font-style: normal;
      font-weight: 600;
      font-display: swap;
      src: url('/fonts/Geist-latin-600.woff2') format('woff2');
    }
    @font-face {
      font-family: 'Geist';
      font-style: normal;
      font-weight: 200;
      font-display: swap;
      src: url('/fonts/Geist-latin-400.woff2') format('woff2');
    }
  </style>
</svelte:head>

<a class="skip" href="#main">Langsung ke konten</a>

<header class="nav">
  <div class="nav-inner">
    <a href={resolve('/')} class="brand">AlgoHub.</a>
    <a href={resolve('/contact')} class="contact">Get in touch</a>
  </div>
</header>

<div class="wrap">
  <main id="main" class="main">
    {@render children()}
  </main>

  <footer class="footer">
    <div class="panel">
      <div class="inner">
        <p class="lead" aria-label="Satu ekosistem untuk kemudahan akademik.">
          {#each 'Satu ekosistem untuk kemudahan akademik.'.split(' ') as word, i (i)}<span
              class="w"
              style="--i: {i}">{word}</span
            >{/each}
        </p>
        <div class="bar">
          <p class="copyright">© {new Date().getFullYear()} AlgoHub · Lab AP</p>
          <nav class="links" aria-label="Layanan">
            {#each services as service (service.href)}
              <a href={service.href} class="link" target="_blank" rel="noreferrer">{service.name}</a
              >
            {/each}
          </nav>
        </div>
        <div class="wordmark-wrap">
          <p class="wordmark" aria-label="AlgoHub">
            {#each 'AlgoHub'.split('') as letter, i (i)}<span style="--i: {i}">{letter}</span
              >{/each}
          </p>
        </div>
      </div>
    </div>
  </footer>
</div>

<nav class="dock" aria-label="Layanan cepat">
  <div class="dock-inner">
    {#each dock as item, i (item.key)}
      {#if i > 0}
        <div class="sep-wrap"><span class="sep"></span></div>
      {/if}
      <a class="dock-item" href={item.href} target="_blank" rel="noreferrer" aria-label={item.name}>
        <span class="tip">
          {item.name}
          <span class="key">{item.key}</span>
        </span>
        <span class="surf">
          <item.icon size={18} strokeWidth={1.5} aria-hidden="true" />
        </span>
      </a>
    {/each}
    <span class="zoom">100%</span>
  </div>
</nav>

<style>
  /* ---------- Global ground ---------- */
  :global(body) {
    background-color: #f4f4f4;
    color: #0a0a0a;
    font-family: 'Geist', 'Inter', ui-sans-serif, system-ui, sans-serif;
  }

  .skip {
    position: absolute;
    left: -9999px;
    top: 0;
    z-index: 100;
    background: #0a0a0a;
    color: #f4f4f4;
    padding: 10px 16px;
    border-radius: 0 0 8px 0;
    text-decoration: none;
    font-size: 13px;
    font-weight: 500;
  }
  .skip:focus {
    left: 0;
  }

  /* ---------- Navbar ---------- */
  .nav {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    z-index: 5;
  }
  .nav-inner {
    display: flex;
    width: 100%;
    padding: 16px 88px 0;
    justify-content: space-between;
    align-items: center;
  }
  .brand {
    color: #0a0a0a;
    font-family: 'Geist', sans-serif;
    font-size: 20px;
    font-weight: 600;
    letter-spacing: -0.03em;
    line-height: 1.25em;
    text-decoration: none;
  }
  .contact {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    padding: 14px 26px;
    background-color: #0a0a0a;
    border-radius: 999px;
    color: #f4f4f4;
    font-family: 'Geist', sans-serif;
    font-size: 13px;
    font-weight: 500;
    line-height: 1.3em;
    text-decoration: none;
    transition:
      opacity 0.2s ease,
      transform 0.2s ease;
  }
  .contact:hover {
    opacity: 0.85;
    transform: translateY(-1px);
  }

  /* ---------- Layout wrapper ---------- */
  .wrap {
    display: flex;
    flex-flow: column;
    min-height: 100vh;
  }
  .main {
    display: flex;
    flex: 1 1 0;
    flex-flow: column;
  }

  /* ---------- Footer ---------- */
  .footer {
    padding: 8px 8px 90px;
  }
  .panel {
    background-color: #0a0a0a;
    border-radius: 32px;
    padding: 110px 88px 28px;
  }
  .inner {
    max-width: 1600px;
    display: flex;
    flex-flow: column;
    margin: 0 auto;
  }
  .lead {
    display: block;
    color: #f4f4f4;
    font-family: 'Geist', sans-serif;
    font-size: 78px;
    font-weight: 600;
    letter-spacing: -0.045em;
    line-height: 1.08em;
    margin: 0;
  }
  .lead .w {
    display: inline-block;
  }
  .bar {
    display: flex;
    flex-wrap: wrap;
    justify-content: space-between;
    align-items: center;
    gap: 20px;
    padding: 80px 0 0;
  }
  .copyright {
    color: #8a8a8a;
    font-family: 'Geist', sans-serif;
    font-size: 13px;
    font-weight: 500;
    line-height: 1.3em;
    margin: 0;
  }
  .links {
    display: flex;
    flex-flow: row;
    gap: 20px;
  }
  .link {
    color: #cfcfcf;
    font-family: 'Geist', sans-serif;
    font-size: 14px;
    font-weight: 400;
    line-height: 1.5em;
    text-decoration: none;
    transition: color 0.2s ease;
  }
  .link:hover {
    color: #8a1538;
  }
  .wordmark-wrap {
    display: flex;
    flex-flow: column;
    align-items: center;
    padding: 30px 0 0;
  }
  .wordmark {
    display: flex;
    justify-content: center;
    white-space: nowrap;
    user-select: none;
    width: 100%;
    color: #2a2a2a;
    font-family: 'Geist', sans-serif;
    font-size: 174px;
    letter-spacing: -0.05em;
    line-height: 0.82em;
    font-weight: 200;
    margin: 0;
  }
  .wordmark span {
    display: inline-block;
    transition: color 0.4s;
  }

  @media (prefers-reduced-motion: no-preference) {
    .wordmark span {
      animation: wave 3.5s ease-in-out infinite;
      animation-delay: calc(var(--i) * 0.12s);
    }
  }
  @keyframes wave {
    0%,
    100% {
      color: #2a2a2a;
    }
    50% {
      color: #8a1538;
    }
  }

  /* ---------- Dock ---------- */
  .dock {
    position: fixed;
    bottom: 18px;
    left: 0;
    right: 0;
    z-index: 5;
    display: flex;
    place-content: center;
    align-items: center;
    pointer-events: none;
  }
  .dock-inner {
    display: flex;
    place-content: center;
    align-items: center;
    gap: 3px;
    padding: 6px;
    backdrop-filter: blur(14px);
    background-color: rgba(255, 255, 255, 0.92);
    border: 1px solid rgba(10, 10, 10, 0.1);
    border-radius: 16px;
    box-shadow: 0px 14px 36px -16px rgba(0, 0, 0, 0.38);
    transition:
      translate 0.6s cubic-bezier(0.16, 1, 0.3, 1),
      opacity 0.3s;
    pointer-events: auto;
  }
  .dock-item {
    position: relative;
    display: flex;
    width: 42px;
    height: 42px;
    align-items: center;
    justify-content: center;
    text-decoration: none;
  }
  .dock-item .surf {
    display: flex;
    width: 100%;
    height: 100%;
    align-items: center;
    justify-content: center;
    border-radius: 11px;
    color: #0a0a0a;
    transition: background-color 0.2s ease;
  }
  .dock-item:hover .surf {
    background-color: rgba(10, 10, 10, 0.06);
  }
  .tip {
    position: absolute;
    bottom: 54px;
    left: 0;
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 6px 9px;
    background-color: #0a0a0a;
    border-radius: 7px;
    color: #ffffff;
    font-family: 'Geist', sans-serif;
    font-size: 12px;
    font-weight: 400;
    line-height: 1.2em;
    white-space: nowrap;
    opacity: 0;
    transform: translateY(6px);
    transition:
      opacity 0.2s ease,
      transform 0.2s ease;
    pointer-events: none;
  }
  .dock-item:hover .tip,
  .dock-item:focus-visible .tip {
    opacity: 1;
    transform: translateY(0);
  }
  .key {
    display: inline-flex;
    align-items: center;
    padding: 0 5px;
    border: 1px solid rgba(255, 255, 255, 0.2);
    border-radius: 4px;
    color: #d6ff3f;
    font-family: 'Geist', sans-serif;
    font-size: 11px;
    font-weight: 500;
    line-height: 1.4em;
  }
  .sep-wrap {
    display: flex;
    padding: 0 5px;
  }
  .sep {
    display: block;
    width: 1px;
    height: 24px;
    background-color: rgba(10, 10, 10, 0.1);
  }
  .zoom {
    display: flex;
    width: 50px;
    align-items: center;
    justify-content: center;
    color: #605f5f;
    font-family: 'Geist', sans-serif;
    font-size: 12px;
    font-weight: 400;
  }

  /* ---------- Breakpoints ---------- */
  @media (max-width: 1439.98px) {
    .nav-inner {
      padding: 16px 72px 0;
    }
    .brand {
      font-size: 19px;
    }
    .contact {
      padding: 11px 20px;
    }
    .panel {
      padding: 96px 72px 28px;
    }
    .lead {
      font-size: 65px;
    }
    .wordmark {
      font-size: 209px;
    }
  }
  @media (max-width: 1199.98px) {
    .nav-inner {
      padding: 16px 48px 0;
    }
    .panel {
      padding: 65px 48px 28px;
    }
    .lead {
      font-size: 44px;
    }
    .wordmark {
      font-size: 117px;
    }
  }
  @media (max-width: 809.98px) {
    .contact {
      padding: 10px 18px;
    }
    .panel {
      padding: 65px 48px 28px;
    }
    .lead {
      font-size: 36px;
    }
    .bar {
      flex-flow: column;
      align-items: flex-start;
      padding: 30px 0 0;
    }
    .wordmark {
      font-size: 57px;
    }
    .dock {
      bottom: 60px;
    }
  }
</style>
