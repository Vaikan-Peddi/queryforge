<script lang="ts">
  import { goto } from '$app/navigation';
  import { FolderPlus } from 'lucide-svelte';
  import AppShell from '$lib/components/AppShell.svelte';
  import { api } from '$lib/api';

  let name = '';
  let error = '';
  let loading = false;

  async function submit() {
    loading = true;
    error = '';
    try {
      const workspace = await api.createWorkspace(name);
      goto(`/workspaces/${workspace.id}`);
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to create workspace';
    } finally {
      loading = false;
    }
  }
</script>

<AppShell>
  <main class="mx-auto max-w-2xl px-4 py-8">
    <section class="rounded-lg border border-line bg-white p-6 shadow-soft">
      <div class="mb-6 flex items-center gap-3">
        <div class="grid h-10 w-10 place-items-center rounded-md bg-accent text-white"><FolderPlus size={20} /></div>
        <div>
          <h1 class="text-xl font-semibold">Create Workspace</h1>
          <p class="text-sm text-slate-500">Name the workspace, then upload a SQLite database.</p>
        </div>
      </div>
      <form class="space-y-4" on:submit|preventDefault={submit}>
        <label class="block text-sm font-medium">Workspace name<input class="focus-ring mt-1 w-full rounded-md border border-line px-3 py-2" bind:value={name} required /></label>
        {#if error}<p class="rounded-md bg-red-50 px-3 py-2 text-sm text-red-700">{error}</p>{/if}
        <button class="focus-ring rounded-md bg-ink px-4 py-2 font-medium text-white disabled:opacity-60" disabled={loading}>{loading ? 'Creating' : 'Create workspace'}</button>
      </form>
    </section>
  </main>
</AppShell>
