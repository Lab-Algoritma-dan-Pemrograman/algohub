<script lang="ts">
  const services = [
    { name: 'SIAKAD', href: 'https://siakad.algohub.web.id' },
    { name: 'e-Learning', href: 'https://elearning.algohub.web.id' },
    { name: 'Praktikum', href: 'https://praktikum.algohub.web.id' },
  ];
</script>

<svelte:head>
  <title>Get in touch · AlgoHub</title>
</svelte:head>

<section class="contact-page">
  <div class="inner">
    <h1 class="title">Get in touch</h1>
    <p class="sub">
      Hubungi tim AlgoHub untuk pertanyaan seputar SIAKAD, e-Learning, atau Praktikum.
    </p>
    <ul class="services">
      {#each services as s (s.href)}
        <li>
          <a href={s.href} target="_blank" rel="noreferrer">{s.name}</a>
        </li>
      {/each}
    </ul>
  </div>
</section>

<style>
  .contact-page {
    padding: 140px 72px 80px;
  }
  .inner {
    max-width: 1776px;
    margin: 0 auto;
    display: flex;
    flex-flow: column;
    gap: 18px;
  }
  .title {
    font-family: 'Geist', sans-serif;
    font-size: 78px;
    font-weight: 600;
    letter-spacing: -0.045em;
    line-height: 1.08em;
    color: #0a0a0a;
    margin: 0;
  }
  .sub {
    color: #605f5f;
    font-family: 'Geist', sans-serif;
    font-size: 16px;
    font-weight: 400;
    line-height: 1.55em;
    max-width: 480px;
    margin: 0;
  }
  .services {
    display: flex;
    flex-wrap: wrap;
    gap: 12px;
    padding: 8px 0 0;
    list-style: none;
    margin: 0;
  }
  .services a {
    display: inline-flex;
    padding: 14px 26px;
    border-radius: 999px;
    background-color: #0a0a0a;
    color: #f4f4f4;
    font-family: 'Geist', sans-serif;
    font-size: 14px;
    font-weight: 500;
    letter-spacing: -0.01em;
    line-height: 1.5em;
    text-decoration: none;
  }
  @media (max-width: 809.98px) {
    .contact-page {
      padding: 110px 24px 60px;
    }
    .title {
      font-size: 36px;
    }
  }
</style>
