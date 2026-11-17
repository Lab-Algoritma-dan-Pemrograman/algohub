<script lang="ts">
  import Binary from '@lucide/svelte/icons/binary';
  import Braces from '@lucide/svelte/icons/braces';
  import Code from '@lucide/svelte/icons/code';
  import Cpu from '@lucide/svelte/icons/cpu';
  import Hash from '@lucide/svelte/icons/hash';
  import Layers from '@lucide/svelte/icons/layers';
  import ScanLine from '@lucide/svelte/icons/scan-line';
  import Terminal from '@lucide/svelte/icons/terminal';
</script>

<div class="pointer-events-none absolute inset-0 z-0 overflow-hidden select-none">
  <div class="animate-float absolute top-[10%] left-[8%] opacity-70 md:opacity-100">
    <div
      class="flex rotate-12 transform items-center gap-2 rounded-2xl border border-white/40 bg-white/60 p-3 shadow-md backdrop-blur-md transition-transform hover:scale-110"
    >
      <Terminal class="text-primary h-6 w-6 shadow-sm" />
      <span class="text-primary hidden font-mono text-sm font-black sm:inline">&gt;_</span>
    </div>
  </div>
  <div class="animate-float-delayed absolute top-[18%] right-[12%] opacity-70 md:opacity-100">
    <div
      class="bg-accent/40 border-accent/30 flex -rotate-6 transform items-center gap-2 rounded-full border px-4 py-2 shadow-md backdrop-blur-md transition-transform hover:scale-110"
    >
      <ScanLine class="text-accent-foreground h-5 w-5" />
      <span class="text-accent-foreground font-mono text-sm font-black">scanf("%d")</span>
    </div>
  </div>
  <div class="animate-float-delayed absolute bottom-[25%] left-[10%] opacity-60 md:opacity-100">
    <div
      class="bg-primary/10 border-primary/30 -rotate-12 transform rounded-xl border p-4 shadow-md backdrop-blur-sm transition-transform hover:scale-110"
    >
      <Braces class="text-primary h-8 w-8 font-bold" />
    </div>
  </div>
  <div class="animate-float absolute top-[35%] left-[22%] opacity-40 md:opacity-80">
    <div
      class="bg-secondary/60 border-secondary/40 rotate-45 transform rounded-lg border px-3 py-1 shadow-sm backdrop-blur-sm"
    >
      <span class="text-secondary-foreground font-mono text-2xl font-black">&amp;</span>
    </div>
  </div>
  <div class="animate-float absolute top-[8%] right-[40%] opacity-50 md:opacity-90">
    <div
      class="-rotate-2 transform rounded-lg border border-indigo-500/30 bg-indigo-500/20 px-4 py-1.5 shadow-sm backdrop-blur-sm"
    >
      <span class="truncate font-mono text-xs font-bold text-indigo-600">for(i=0; i&lt;n; i++)</span
      >
    </div>
  </div>
  <div class="animate-float absolute top-[5%] right-[25%] opacity-50 md:opacity-90">
    <div
      class="-rotate-3 transform rounded-2xl border border-white/30 bg-white/80 p-3 shadow-md backdrop-blur-md transition-transform hover:scale-110"
    >
      <Code class="h-10 w-10 text-slate-500" />
    </div>
  </div>
  <div class="animate-float absolute bottom-[30%] left-[30%] opacity-40 md:opacity-70">
    <div
      class="rotate-12 transform rounded-xl border border-orange-500/20 bg-orange-500/10 p-3 shadow-sm backdrop-blur-sm"
    >
      <Cpu class="h-7 w-7 text-orange-600" />
    </div>
  </div>
  <div class="animate-float absolute right-[8%] bottom-[30%] opacity-70 md:opacity-100">
    <div
      class="flex rotate-2 transform items-center gap-2 rounded-2xl border border-slate-300 bg-white/70 px-5 py-2 shadow-md backdrop-blur-md transition-transform hover:scale-110"
    >
      <span class="font-mono text-sm font-black text-slate-800">printf("Lab AP")</span>
    </div>
  </div>
  <div class="animate-float-delayed absolute top-[55%] left-[12%] opacity-50 md:opacity-90">
    <div
      class="rounded-md border border-emerald-500/30 bg-emerald-500/20 px-3 py-1 shadow-sm backdrop-blur-sm"
    >
      <span class="font-mono text-xs font-bold text-emerald-700">if(isValid)</span>
    </div>
  </div>
  <div class="animate-float-delayed absolute right-[22%] bottom-[8%] opacity-40 md:opacity-70">
    <div
      class="bg-primary/20 border-primary/20 rotate-12 transform rounded-full border p-3 shadow-md backdrop-blur-sm transition-transform hover:scale-110"
    >
      <Binary class="text-primary h-6 w-6" />
    </div>
  </div>
  <div class="animate-float absolute bottom-[10%] left-[25%] opacity-40 md:opacity-70">
    <div
      class="-rotate-12 transform rounded-xl border border-blue-500/20 bg-blue-500/10 p-3 shadow-sm backdrop-blur-sm"
    >
      <Layers class="h-6 w-6 text-blue-600" />
    </div>
  </div>
  <div class="animate-float-delayed absolute top-[65%] right-[18%] opacity-50 md:opacity-80">
    <div
      class="bg-accent/30 border-accent/20 flex -rotate-12 transform items-center gap-2 rounded-xl border px-4 py-2 shadow-md backdrop-blur-md transition-transform hover:scale-110"
    >
      <Hash class="text-accent-foreground h-4 w-4" />
      <span class="text-accent-foreground font-mono text-sm font-black">while(1)</span>
    </div>
  </div>
  <div
    class="bg-primary/10 animate-pulse-soft absolute top-1/4 left-1/3 -z-10 h-64 w-64 rounded-full blur-3xl"
  ></div>
  <div
    class="bg-accent/10 animate-pulse-soft absolute right-1/4 bottom-1/4 -z-10 h-80 w-80 rounded-full blur-3xl"
  ></div>
</div>
