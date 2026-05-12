<script lang="ts">
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { Bot, Database, FolderOpen, Plus, Trash2 } from 'lucide-svelte';
  import AppShell from '$lib/components/AppShell.svelte';
  import { api, getSession } from '$lib/api';
  import type { AIHealth, Workspace } from '$lib/types';

  let workspaces: Workspace[] = [];
  let aiHealth: AIHealth | null = null;
  let aiHealthError = '';
  let error = '';
  let loading = true;

  onMount(async () => {
    if (!getSession()) return goto('/login');
    try {
      const [workspaceResponse, healthResponse] = await Promise.allSettled([api.workspaces(), api.aiHealth()]);
      if (workspaceResponse.status === 'fulfilled') {
        workspaces = workspaceResponse.value.workspaces;
      } else {
        throw workspaceResponse.reason;
      }
      if (healthResponse.status === 'fulfilled') {
        aiHealth = healthResponse.value;
      } else {
        aiHealthError = healthResponse.reason instanceof Error ? healthResponse.reason.message : 'AI service unavailable';
      }
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to load workspaces';
    } finally {
      loading = false;
    }
  });

  async function remove(id: string) {
    if (!confirm('Delete this workspace and its history?')) return;
    await api.deleteWorkspace(id);
    workspaces = workspaces.filter((w) => w.id !== id);
  }
</script>

<AppShell>
  <main class="mx-auto max-w-7xl px-4 py-8">
    <div class="mb-6 flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-semibold">Workspaces</h1>
        <p class="text-sm text-slate-500">SQLite databases ready for inspection and analysis.</p>
      </div>
      <a class="focus-ring flex items-center gap-2 rounded-md bg-ink px-4 py-2 text-sm font-medium text-white" href="/workspaces/new"><Plus size={17} /> Create</a>
    </div>
    {#if error}<p class="mb-4 rounded-md bg-red-50 px-3 py-2 text-sm text-red-700">{error}</p>{/if}
    <section class="mb-6 rounded-lg border border-line bg-white p-4 shadow-soft">
      <div class="flex flex-wrap items-center justify-between gap-3">
        <div class="flex items-center gap-3">
          <div class="grid h-10 w-10 place-items-center rounded-md bg-panel text-accent"><Bot size={19} /></div>
          <div>
            <h2 class="font-semibold">LLM Provider</h2>
            <p class="text-sm text-slate-500">
              {#if aiHealth}
                {aiHealth.provider} · {aiHealth.model}
              {:else if aiHealthError}
                {aiHealthError}
              {:else}
                Checking AI service...
              {/if}
            </p>
          </div>
        </div>
        {#if aiHealth}
          <span class="rounded-md bg-emerald-50 px-3 py-1 text-sm font-medium text-emerald-700">{aiHealth.status}</span>
        {:else}
          <span class="rounded-md bg-amber-50 px-3 py-1 text-sm font-medium text-amber-700">unavailable</span>
        {/if}
      </div>
    </section>
    {#if loading}
      <div class="rounded-lg border border-line bg-white p-8 text-sm text-slate-500">Loading workspaces...</div>
    {:else if workspaces.length === 0}
      <section class="rounded-lg border border-line bg-white p-10 text-center shadow-soft">
        <Database class="mx-auto mb-4 text-accent" size={40} />
        <h2 class="text-lg font-semibold">No workspaces yet</h2>
        <p class="mt-1 text-sm text-slate-500">Create one and upload a SQLite database to begin.</p>
      </section>
    {:else}
      <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
        {#each workspaces as workspace}
          <article class="rounded-lg border border-line bg-white p-5 shadow-soft">
            <div class="mb-4 flex items-start justify-between gap-4">
              <div>
                <h2 class="font-semibold">{workspace.name}</h2>
                <p class="text-sm text-slate-500">{workspace.db_type.toUpperCase()} · {workspace.has_database ? 'database uploaded' : 'waiting for upload'}</p>
              </div>
              <button class="focus-ring grid h-8 w-8 place-items-center rounded-md border border-line text-slate-500" title="Delete workspace" aria-label="Delete workspace" on:click={() => remove(workspace.id)}><Trash2 size={15} /></button>
            </div>
            <a class="focus-ring flex items-center justify-center gap-2 rounded-md bg-accent px-3 py-2 text-sm font-medium text-white" href={`/workspaces/${workspace.id}`}><FolderOpen size={16} /> Open workspace</a>
          </article>
        {/each}
      </div>
    {/if}
  </main>
</AppShell>
