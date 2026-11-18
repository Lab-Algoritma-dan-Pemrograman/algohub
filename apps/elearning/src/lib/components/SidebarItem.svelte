<script lang="ts">
  import { resolve } from '$app/paths';
  import type { RouteId } from '$app/types';
  import type { Component } from 'svelte';

  let {
    href,
    label,
    icon: Icon,
    active = false,
    hoverClass = 'group-hover:text-maroon',
  }: {
    href: string;
    label: string;
    icon: Component<{ size?: number }>;
    active?: boolean;
    hoverClass?: string;
  } = $props();
</script>

<a
  href={resolve(href as RouteId)}
  aria-current={active ? 'page' : undefined}
  class="group relative flex items-center space-x-4 overflow-hidden rounded-2xl px-4 py-3 font-bold transition-all {active
    ? 'bg-maroon shadow-bubbly-maroon hover:-translate-y-1 font-black text-white'
    : 'hover:bg-maroon-bg hover:text-maroon text-gray-500'}"
>
  <div
    class="flex h-9 w-9 items-center justify-center rounded-xl transition-all {active
      ? 'relative z-10 bg-white/20 text-white'
      : `bg-gray-100 text-gray-400 group-hover:bg-white group-hover:shadow-sm ${hoverClass}`}"
  >
    <Icon size={16} />
  </div>
  <span class="relative z-10 text-sm font-bold tracking-wide {active ? 'font-black' : ''}"
    >{label}</span
  >
</a>
