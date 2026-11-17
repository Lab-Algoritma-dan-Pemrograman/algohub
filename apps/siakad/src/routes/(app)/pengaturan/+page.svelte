<script lang="ts">
  import {
    Badge,
    Button,
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle,
    Input,
    Label,
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    Switch,
    Textarea,
  } from '@algohub/ui';
  import BookOpen from '@lucide/svelte/icons/book-open';
  import CalendarClock from '@lucide/svelte/icons/calendar-clock';
  import Info from '@lucide/svelte/icons/info';
  import ListChecks from '@lucide/svelte/icons/list-checks';
  import Save from '@lucide/svelte/icons/save';
  import Send from '@lucide/svelte/icons/send';
  import Settings from '@lucide/svelte/icons/settings';
  import Smartphone from '@lucide/svelte/icons/smartphone';
  import { createQuery } from '@tanstack/svelte-query';
  import { triggerServerWhatsApp } from '#lib/notifications.ts';
  import { createRpcMutation, rpcQueryOptions } from '#lib/queries/rpc.ts';
  import { getDailyQuote } from '#lib/quotes.ts';
  import { toast } from '#lib/toast.ts';
  import type { PageProps } from './$types';

  interface SystemSettings {
    semester_active?: string;
    announcement?: string;
    is_recruitment_open?: boolean;
    recruitment_link?: string;
    login_guide_text?: string;
    procedure_text?: { step: string; desc: string }[] | string | null;
    reschedule_steps?: string;
    wa_gateway_provider?: string;
    wa_gateway_token?: string;
    wa_auto_notify_enabled?: boolean;
    wa_templates?: Record<string, string> | null;
  }

  const PROVIDERS: Record<string, string> = {
    fonnte: 'Fonnte (Recommended - api.fonnte.com)',
    wablas: 'Wablas (kudus.wablas.com)',
    whacenter: 'Whacenter (app.whacenter.com)',
    custom: 'Custom Webhook HTTP API',
  };

  const PROCEDURE_PLACEHOLDER =
    'Pendaftaran Akun: Daftarkan akun menggunakan NIM/ID yang valid.\nMasuk Sistem: Login ke portal dengan NIM dan password saat shift aktif.\nPresensi Kehadiran: Lakukan presensi di menu kehadiran tepat waktu.\nModul & Peminjaman: Unduh modul di menu E-Learning atau ajukan sewa barang lab.';
  const RESCHEDULE_PLACEHOLDER =
    'Ajukan Izin: Klik menu Absensi setelah login, lalu ajukan izin di tab Izin.\nKirim WhatsApp: Klik Kirim Bukti WA untuk membagikan berkas pendukung ke Asisten.\nVerifikasi: Tunggu hingga status pengajuan izin berubah menjadi terverifikasi.\nPilih Slot Baru: Klik Pilih Jadwal dan tentukan slot ganti yang kosong.\nACC Asisten: Tunggu persetujuan jadwal pengganti agar presensi dianggap sah.';

  let { data }: PageProps = $props();
  const user = $derived(data.auth.user);

  const settingsQuery = createQuery(() =>
    rpcQueryOptions<SystemSettings[]>('get_system_settings_full_secure', { p_viewer_id: user.id }),
  );
  const updateSettings = createRpcMutation('admin_update_global_settings_secure');

  let settings = $state<SystemSettings>({});
  let loading = $state(false);
  let loaded = false;

  $effect(() => {
    const rows = settingsQuery.data;
    if (loaded || !rows || rows.length === 0) return;
    loaded = true;
    settings = { ...rows[0] };
  });

  async function handleSave() {
    if (user.role !== 'koordinator') {
      toast.error('Akses Ditolak', {
        description: 'Hanya Koordinator yang punya kunci ke pengaturan sistem.',
      });
      return;
    }
    loading = true;
    let procedureValue = settings.procedure_text;
    if (typeof procedureValue === 'string') {
      try {
        procedureValue = JSON.parse(procedureValue);
      } catch {
        procedureValue = null;
      }
    }
    const { error } = await updateSettings.mutateAsync({
      p_caller_id: user.id,
      p_semester_active: settings.semester_active || '',
      p_announcement: settings.announcement || '',
      p_is_recruitment_open: settings.is_recruitment_open || false,
      p_wa_templates: settings.wa_templates || null,
      p_recruitment_link: settings.recruitment_link || '',
      p_login_guide_text: settings.login_guide_text || '',
      p_procedure_text: procedureValue || null,
      p_reschedule_steps: settings.reschedule_steps || '',
      p_wa_gateway_provider: settings.wa_gateway_provider || 'fonnte',
      p_wa_gateway_token: settings.wa_gateway_token || '',
      p_wa_auto_notify_enabled: settings.wa_auto_notify_enabled !== false,
    });
    if (error) toast.error('Gagal menyimpan pengaturan: ' + error.message);
    else toast.success('Pengaturan diperbarui!');
    loading = false;
  }

  const procedureAsText = $derived.by(() => {
    const p = settings.procedure_text;
    if (!p) return '';
    if (Array.isArray(p)) return p.map((item) => `${item.step}: ${item.desc}`).join('\n');
    return typeof p === 'string' ? p : JSON.stringify(p, null, 2);
  });

  function handleProcedureChange(text: string) {
    settings.procedure_text = text
      .split('\n')
      .filter((l) => l.trim())
      .map((line) => {
        const idx = line.indexOf(':');
        return idx > -1
          ? { step: line.substring(0, idx).trim(), desc: line.substring(idx + 1).trim() }
          : { step: line.trim(), desc: '' };
      });
  }

  function setTemplate(key: string, value: string) {
    settings.wa_templates = { ...settings.wa_templates, [key]: value };
  }

  async function handleTestWa() {
    const quote = getDailyQuote('asisten');
    toast.info('📲 Memicu Uji Kirim WA Otomatis dari Server...');
    await triggerServerWhatsApp(
      user.id,
      `🤖 *Tes WA Gateway Lab AP (Server Auto Broadcast)*\n\nHore! Integrasi WhatsApp Gateway dari server Vercel berhasil terhubung!\n\n🌐 *Buka Web:* https://www.lab-ap.web.id\n\n✨ "${quote}"`,
    );
  }
</script>

{#snippet templateField(
  key: string,
  label: string,
  badge: string,
  badgeClass: string,
  placeholder: string,
)}
  <div class="space-y-2">
    <div class="flex items-center justify-between">
      <Label for={`tpl-${key}`}>{label}</Label>
      <Badge variant="outline" class="text-[10px] {badgeClass}">{badge}</Badge>
    </div>
    <Textarea
      id={`tpl-${key}`}
      rows={3}
      value={settings.wa_templates?.[key] || ''}
      oninput={(e) => setTemplate(key, e.currentTarget.value)}
      {placeholder}
    />
  </div>
{/snippet}

<div class="mx-auto max-w-2xl space-y-6">
  <h1 class="flex items-center gap-2 text-2xl font-bold">
    <Settings class="text-primary" /> Pengaturan Sistem
  </h1>

  <Card>
    <CardHeader>
      <CardTitle>Konfigurasi Umum</CardTitle>
      <CardDescription
        >Pengaturan ini akan berdampak pada seluruh pengguna aplikasi.</CardDescription
      >
    </CardHeader>
    <CardContent class="space-y-4">
      <div class="space-y-2">
        <Label for="semester-active">Semester Aktif</Label>
        <Input
          id="semester-active"
          value={settings.semester_active || ''}
          oninput={(e) => (settings.semester_active = e.currentTarget.value)}
          placeholder="Contoh: Genap 2025/2026"
        />
      </div>
      <div class="space-y-2">
        <Label for="announcement">Pengumuman Berjalan (Tampil di Marquee Login)</Label>
        <Textarea
          id="announcement"
          rows={3}
          value={settings.announcement || ''}
          oninput={(e) => (settings.announcement = e.currentTarget.value)}
          placeholder="Tulis pengumuman singkat yang akan berjalan di papan info halaman login..."
        />
        <p class="text-muted-foreground text-[10px]">
          Teks ini tampil di papan marquee berjalan halaman login.
        </p>
      </div>
      <div class="flex items-center justify-between rounded-lg border p-4">
        <div class="space-y-0.5">
          <Label for="recruitment-open" class="text-base">Buka Pendaftaran Asisten?</Label>
          <p class="text-muted-foreground text-sm">Aktifkan jika sedang masa Open Recruitment.</p>
        </div>
        <Switch
          id="recruitment-open"
          checked={settings.is_recruitment_open || false}
          onCheckedChange={(checked) => (settings.is_recruitment_open = checked)}
        />
      </div>
      {#if settings.is_recruitment_open}
        <div
          class="animate-in fade-in space-y-2 rounded-lg border border-pink-100 bg-pink-50/20 p-4 duration-200"
        >
          <Label for="recruitment-link">Tautan Pendaftaran (Oprec Link)</Label>
          <Input
            id="recruitment-link"
            value={settings.recruitment_link || ''}
            oninput={(e) => (settings.recruitment_link = e.currentTarget.value)}
            placeholder="Masukkan link pendaftaran (misal Google Form / WhatsApp)"
          />
          <p class="text-muted-foreground text-[10px]">
            Link ini akan muncul di halaman login pada papan pengumuman Oprec.
          </p>
        </div>
      {/if}
    </CardContent>
  </Card>

  <Card>
    <CardHeader>
      <CardTitle class="flex items-center gap-2 text-blue-700">
        <BookOpen class="h-5 w-5" /> Panduan Login (Modal)
      </CardTitle>
      <CardDescription>
        Teks yang tampil di dalam modal <strong>"Panduan Login"</strong> saat diklik di halaman login.
      </CardDescription>
    </CardHeader>
    <CardContent class="space-y-3">
      <div class="space-y-2">
        <Label for="login-guide">Isi Panduan Login</Label>
        <Textarea
          id="login-guide"
          rows={5}
          value={settings.login_guide_text || ''}
          oninput={(e) => (settings.login_guide_text = e.currentTarget.value)}
          placeholder="Pastikan Anda hanya mencoba login ketika sesi shift kelas Anda sedang aktif. Jika Anda menghadapi kendala login darurat, hubungi Asisten PJ kelas Anda."
        />
        <p class="text-muted-foreground text-[10px]">
          Teks bebas. Ditampilkan sebagai paragraf isi pada modal Panduan Login di halaman login.
        </p>
      </div>
    </CardContent>
  </Card>

  <Card>
    <CardHeader>
      <CardTitle class="flex items-center gap-2 text-emerald-700">
        <ListChecks class="h-5 w-5" /> Alur Prosedur Praktikan (Modal)
      </CardTitle>
      <CardDescription>
        Langkah-langkah yang tampil di modal <strong>"Alur Prosedur"</strong>. Format per baris:
        <code class="rounded bg-slate-100 px-1 text-xs">Judul Langkah: Deskripsi</code>
      </CardDescription>
    </CardHeader>
    <CardContent class="space-y-3">
      <div class="space-y-2">
        <Label for="procedure">Langkah-Langkah Prosedur</Label>
        <Textarea
          id="procedure"
          rows={7}
          value={procedureAsText}
          oninput={(e) => handleProcedureChange(e.currentTarget.value)}
          placeholder={PROCEDURE_PLACEHOLDER}
        />
        <div class="flex items-start gap-1.5 rounded-lg bg-blue-50 p-2 text-[10px] text-blue-700">
          <Info class="mt-0.5 h-3 w-3 shrink-0" />
          <span
            >Format tiap baris: <b>Judul Langkah: Deskripsi singkat.</b> Gunakan Enter untuk memisahkan
            antar langkah.</span
          >
        </div>
      </div>
    </CardContent>
  </Card>

  <Card>
    <CardHeader>
      <CardTitle class="flex items-center gap-2 text-orange-700">
        <CalendarClock class="h-5 w-5" /> Langkah Reschedule (Modal)
      </CardTitle>
      <CardDescription>
        Panduan yang tampil di modal <strong>"Langkah Reschedule"</strong> di halaman login.
      </CardDescription>
    </CardHeader>
    <CardContent class="space-y-3">
      <div class="space-y-2">
        <Label for="reschedule-steps">Isi Panduan Reschedule</Label>
        <Textarea
          id="reschedule-steps"
          rows={6}
          value={settings.reschedule_steps || ''}
          oninput={(e) => (settings.reschedule_steps = e.currentTarget.value)}
          placeholder={RESCHEDULE_PLACEHOLDER}
        />
        <p class="text-muted-foreground text-[10px]">
          Teks bebas multi-baris. Setiap baris menjadi satu butir langkah dalam modal reschedule di
          halaman login.
        </p>
      </div>
    </CardContent>
  </Card>

  <Card class="border-green-200 bg-green-50/30 shadow-sm">
    <CardHeader>
      <CardTitle class="flex items-center gap-2 text-green-800">
        <Send class="h-5 w-5 text-green-600" /> Pengaturan WhatsApp Gateway (Pengiriman Otomatis)
      </CardTitle>
      <CardDescription>
        Hubungkan penyedia WhatsApp Gateway (Fonnte / Wablas / Whacenter) agar pesan pengingat WA
        terkirim 100% otomatis dari server tanpa perlu di-klik.
      </CardDescription>
    </CardHeader>
    <CardContent class="space-y-4">
      <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
        <div class="space-y-2">
          <Label>Penyedia WA Gateway</Label>
          <Select
            type="single"
            value={settings.wa_gateway_provider || 'fonnte'}
            onValueChange={(val) => (settings.wa_gateway_provider = val)}
          >
            <SelectTrigger class="w-full bg-white"
              >{PROVIDERS[settings.wa_gateway_provider || 'fonnte'] ??
                settings.wa_gateway_provider}</SelectTrigger
            >
            <SelectContent>
              {#each Object.entries(PROVIDERS) as [value, label] (value)}<SelectItem
                  {value}
                  {label}
                />{/each}
            </SelectContent>
          </Select>
        </div>
        <div class="space-y-2">
          <Label for="wa-token">API Token WA Gateway</Label>
          <Input
            id="wa-token"
            type="password"
            class="bg-white font-mono text-sm"
            value={settings.wa_gateway_token || ''}
            oninput={(e) => (settings.wa_gateway_token = e.currentTarget.value)}
            placeholder="Contoh: xYz123456789Token..."
          />
        </div>
      </div>
      <div
        class="flex items-center justify-between rounded-lg border border-green-200 bg-white p-3"
      >
        <div>
          <p class="text-sm font-semibold text-green-900">
            Aktifkan Pengiriman WA Otomatis (Auto Broadcast)
          </p>
          <p class="text-muted-foreground text-xs">
            Kirim pengingat jadwal & reschedule otomatis dari server saat ada perubahan
          </p>
        </div>
        <Switch
          aria-label="Aktifkan Pengiriman WA Otomatis"
          checked={settings.wa_auto_notify_enabled !== false}
          onCheckedChange={(val) => (settings.wa_auto_notify_enabled = val)}
        />
      </div>
      <Button
        type="button"
        variant="outline"
        size="sm"
        class="mt-2 border-green-600 bg-white text-green-700 hover:bg-green-50"
        onclick={handleTestWa}
      >
        <Send class="mr-2 h-4 w-4 text-green-600" /> Uji Kirim WA Otomatis dari Server
      </Button>
    </CardContent>
  </Card>

  <Card>
    <CardHeader>
      <CardTitle class="flex items-center gap-2 text-green-700">
        <Smartphone class="h-5 w-5" /> Templat Pesan WhatsApp
      </CardTitle>
      <CardDescription>
        Sesuaikan isi pesan otomatis yang akan dikirim melalui WhatsApp. Gunakan placeholder (nama
        dalam kurung kurawal) untuk data dinamis.
      </CardDescription>
    </CardHeader>
    <CardContent class="space-y-6">
      <div class="space-y-2">
        {@render templateField(
          'absen_izin',
          'Pesan Izin Praktikan (ke Asisten)',
          'Untuk Praktikan',
          '',
          'Contoh: Halo Kak, saya {{nama}} ({{nim}}) ingin izin...',
        )}
        <div class="flex items-start gap-1 rounded bg-blue-50 p-2 text-[10px] text-blue-700">
          <Info class="mt-0.5 h-3 w-3 shrink-0" />
          <span
            >Placeholder: <b
              >&#123;&#123;nama&#125;&#125;, &#123;&#123;nim&#125;&#125;,
              &#123;&#123;kelas&#125;&#125;, &#123;&#123;jurusan&#125;&#125;,
              &#123;&#123;alasan&#125;&#125;</b
            ></span
          >
        </div>
      </div>
      <div class="space-y-2">
        {@render templateField(
          'asisten_swap',
          'Pesan Swap Jadwal (ke Koordinator)',
          'Untuk Asisten',
          'border-orange-200 text-orange-700 bg-orange-50',
          'Contoh: Halo Koordinator, saya {{nama}} ingin swap...',
        )}
        <div class="flex items-start gap-1 rounded bg-blue-50 p-2 text-[10px] text-blue-700">
          <Info class="mt-0.5 h-3 w-3 shrink-0" />
          <span
            >Placeholder: <b
              >&#123;&#123;nama&#125;&#125;, &#123;&#123;jadwal&#125;&#125;,
              &#123;&#123;alasan&#125;&#125;</b
            ></span
          >
        </div>
      </div>
      <div class="space-y-2">
        {@render templateField(
          'chat_asisten',
          'Pesan Chat Asisten (dari Jadwal Saya)',
          'Untuk Praktikan',
          '',
          'Contoh: Selamat {{waktu}} {{panggilan}} {{nama_asisten}}, saya {{nama_praktikan}}...',
        )}
        <div
          class="flex items-start gap-1 rounded border border-green-100 bg-green-50 p-2 text-[10px] text-green-700"
        >
          <Info class="mt-0.5 h-3 w-3 shrink-0" />
          <div class="flex flex-col gap-1">
            <span
              >Placeholder: <b
                >&#123;&#123;waktu&#125;&#125;, &#123;&#123;panggilan&#125;&#125;,
                &#123;&#123;nama_asisten&#125;&#125;, &#123;&#123;nama_praktikan&#125;&#125;,
                &#123;&#123;nim&#125;&#125;, &#123;&#123;jurusan&#125;&#125;,
                &#123;&#123;kelas&#125;&#125;</b
              ></span
            >
            <span class="text-[9px] opacity-70"
              >* waktu: Pagi/Siang/Sore/Malam otomatis. panggilan: Bang/Kak berdasarkan asisten.</span
            >
          </div>
        </div>
      </div>
    </CardContent>
  </Card>

  <div class="pt-2">
    <Button onclick={handleSave} disabled={loading} class="h-12 w-full text-lg shadow-lg">
      <Save class="mr-2 h-5 w-5" />
      {loading ? 'Menyimpan...' : 'Simpan Semua Pengaturan'}
    </Button>
  </div>
</div>
