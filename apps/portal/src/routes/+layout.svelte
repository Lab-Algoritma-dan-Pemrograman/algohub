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
  <footer class="footer">
    <div class="panel">
      <div class="inner">
        <p class="lead">Satu ekosistem untuk kemudahan akademik.</p>
        <div class="bar">
          <p class="copyright">© {new Date().getFullYear()} AlgoHub · Lab AP</p>
          <nav class="links" aria-label="Layanan">
            {#each services as service (service.href)}
              <a href={service.href} class="link">{service.name}</a>
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

<style>
  .footer {
    padding: 8px 8px 24px;
  }
  .panel {
    background-color: #0a0a0a;
    border-radius: 32px;
    padding: 96px 72px 28px;
  }
  .inner {
    max-width: 1600px;
    display: flex;
    flex-flow: column;
    margin: 0 auto;
  }
  .lead {
    color: #f4f4f4;
    font-size: 56px;
    font-weight: 600;
    letter-spacing: -0.045em;
    line-height: 1.08em;
    margin: 0;
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
    transition:
      color 0.4s,
      font-weight 0.4s;
    animation: wave 3.5s ease-in-out infinite;
    animation-delay: calc(var(--i) * 0.12s);
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
      font-weight: 200;
      color: #2a2a2a;
    }
    50% {
      font-weight: 900;
      color: #8a1538;
    }
  }
  @media (max-width: 1199.98px) {
    .panel {
      padding: 65px 48px 28px;
    }
    .lead {
      font-size: 43px;
    }
    .wordmark {
      font-size: 120px;
    }
  }
  @media (max-width: 809.98px) {
    .panel {
      padding: 60px 24px 28px;
    }
    .lead {
      font-size: 34px;
    }
    .bar {
      flex-flow: column;
      align-items: flex-start;
      padding: 40px 0 0;
    }
    .wordmark {
      font-size: 60px;
    }
  }
</style>
