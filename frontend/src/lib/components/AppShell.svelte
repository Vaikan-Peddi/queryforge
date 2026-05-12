<script lang="ts">
  import { goto } from '$app/navigation';
  import { LogOut, Plus } from 'lucide-svelte';
  import { api, clearSession, getSession } from '$lib/api';

  const session = getSession();

  async function logout() {
    const refresh = session?.refresh_token;
    clearSession();
    if (refresh) api.logout(refresh).catch(() => undefined);
    goto('/login');
  }
</script>

<div class="min-h-screen bg-[#eef3f8]">
  <header class="border-b border-line bg-white">
    <div class="mx-auto flex max-w-7xl items-center justify-between px-4 py-3">
      <a href="/dashboard" class="flex items-center gap-3">
        <img class="h-9 w-9 rounded-md" src="/brand/queryforge-logo.svg" alt="QueryForge logo" />
        <div>
          <div class="font-semibold">QueryForge</div>
          <div class="text-xs text-slate-500">{session?.user.email}</div>
        </div>
      </a>
      <div class="flex items-center gap-2">
        <a class="focus-ring flex h-9 items-center gap-2 rounded-md border border-line bg-white px-3 text-sm font-medium" href="/workspaces/new"><Plus size={16} /> New</a>
        <button class="focus-ring grid h-9 w-9 place-items-center rounded-md border border-line bg-white" aria-label="Log out" title="Log out" on:click={logout}><LogOut size={16} /></button>
      </div>
    </div>
  </header>
  <slot />
</div>
