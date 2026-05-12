<script lang="ts">
  import { onMount } from 'svelte';
  import { fade } from 'svelte/transition';
  import { page } from '$app/stores';
  import { goto } from '$app/navigation';
  import { Bot, ChevronRight, Database, FileUp, Play, RefreshCcw, Table2, Trash2, Zap } from 'lucide-svelte';
  import AppShell from '$lib/components/AppShell.svelte';
  import { api, getSession } from '$lib/api';
  import type { DatabaseSchema, HistoryItem, QueryResult, Workspace } from '$lib/types';

  let workspace: Workspace;
  let schema: DatabaseSchema | null = null;
  let history: HistoryItem[] = [];
  let question = 'Show me the first 10 rows';
  let sql = '';
  let explanation = '';
  let confidence = 0;
  let result: QueryResult | null = null;
  let error = '';
  let loading = true;
  let busy = false;
  let generating = false;
  let fileInput: HTMLInputElement;

  const PHRASES = [
    'Reading your schema...',
    'Thinking hard...',
    'Translating to SQL...',
    'Picking the right tables...',
    'Joining the dots...',
    'Writing the query...',
    'Checking your columns...',
    'Pondering foreign keys...',
    'Making it read-only...',
    'Polishing the SQL...',
    'Consulting the oracle...',
    'Squinting at your question...',
    'Crafting the perfect SELECT...',
    'Almost there...',
  ];

  let phraseIdx = 0;
  let phraseTimer: ReturnType<typeof setInterval> | null = null;

  function startPhrases() {
    phraseIdx = Math.floor(Math.random() * PHRASES.length);
    phraseTimer = setInterval(() => {
      phraseIdx = (phraseIdx + 1) % PHRASES.length;
    }, 1800);
  }

  function stopPhrases() {
    if (phraseTimer) { clearInterval(phraseTimer); phraseTimer = null; }
  }

  $: workspaceId = $page.params.id;

  onMount(async () => {
    if (!getSession()) return goto('/login');
    await loadAll();
  });

  async function loadAll() {
    error = '';
    loading = true;
    try {
      workspace = await api.workspace(workspaceId);
      await Promise.all([loadSchema(), loadHistory()]);
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to load workspace';
    } finally {
      loading = false;
    }
  }

  async function loadSchema() {
    if (!workspace?.has_database) return;
    schema = await api.schema(workspaceId);
  }

  async function loadHistory() {
    history = (await api.history(workspaceId)).history || [];
  }

  async function upload() {
    const file = fileInput.files?.[0];
    if (!file) return;
    busy = true;
    error = '';
    try {
      workspace = await api.uploadDatabase(workspaceId, file);
      schema = await api.schema(workspaceId);
      await loadHistory();
    } catch (err) {
      error = err instanceof Error ? err.message : 'Upload failed';
    } finally {
      busy = false;
    }
  }

  async function generate() {
    busy = true;
    generating = true;
    startPhrases();
    error = '';
    try {
      const generated = await api.generate(workspaceId, question);
      sql = generated.sql;
      explanation = generated.explanation;
      confidence = generated.confidence;
      await loadHistory();
    } catch (err) {
      error = err instanceof Error ? err.message : 'Generation failed';
    } finally {
      stopPhrases();
      generating = false;
      busy = false;
    }
  }

  async function execute() {
    busy = true;
    error = '';
    try {
      result = await api.execute(workspaceId, sql);
      await loadHistory();
    } catch (err) {
      error = err instanceof Error ? err.message : 'Execution failed';
    } finally {
      busy = false;
    }
  }

  async function generateAndRun() {
    busy = true;
    generating = true;
    startPhrases();
    error = '';
    try {
      const generated = await api.generate(workspaceId, question);
      sql = generated.sql;
      explanation = generated.explanation;
      confidence = generated.confidence;
      stopPhrases();
      generating = false;
      result = await api.execute(workspaceId, sql);
      await loadHistory();
    } catch (err) {
      error = err instanceof Error ? err.message : 'Generate & run failed';
    } finally {
      stopPhrases();
      generating = false;
      busy = false;
    }
  }

  function clearHistory() {
    history = [];
  }

  function useHistory(item: HistoryItem) {
    question = item.question || question;
    sql = item.executed_sql || item.generated_sql || sql;
    explanation = item.explanation || explanation;
  }
</script>

<AppShell>
  <main class="px-4 py-6">
    {#if loading}
      <div class="rounded-lg border border-line bg-white p-8 text-sm text-slate-500">Loading workspace...</div>
    {:else}
      <div class="mb-4 flex flex-wrap items-center justify-between gap-3">
        <div>
          <h1 class="text-2xl font-semibold">{workspace.name}</h1>
          <p class="text-sm text-slate-500">SQLite analysis workspace</p>
        </div>
        <label class="focus-ring flex cursor-pointer items-center gap-2 rounded-md border border-line bg-white px-3 py-2 text-sm font-medium">
          <FileUp size={16} /> Upload database
          <input class="hidden" type="file" accept=".sqlite,.sqlite3,.db" bind:this={fileInput} on:change={upload} />
        </label>
      </div>
      {#if error}<p class="mb-4 rounded-md bg-red-50 px-3 py-2 text-sm text-red-700">{error}</p>{/if}
      {#if !workspace.has_database}
        <section class="rounded-lg border border-line bg-white p-10 text-center shadow-soft">
          <Database class="mx-auto mb-4 text-accent" size={42} />
          <h2 class="text-lg font-semibold">Upload a SQLite database</h2>
          <p class="mt-1 text-sm text-slate-500">Accepted formats are .sqlite, .sqlite3, and .db.</p>
        </section>
      {:else}
        <div class="grid gap-4 lg:grid-cols-[280px_1fr_300px]">
          <aside class="max-h-[calc(100vh-10rem)] self-start overflow-y-auto rounded-lg border border-line bg-white p-4 shadow-soft">
            <div class="mb-3 flex items-center justify-between">
              <h2 class="flex items-center gap-2 font-semibold"><Table2 size={17} /> Schema</h2>
              <button class="focus-ring grid h-8 w-8 place-items-center rounded-md border border-line" title="Refresh schema" aria-label="Refresh schema" on:click={loadSchema}><RefreshCcw size={14} /></button>
            </div>
            <div class="space-y-3">
              {#each schema?.tables || [] as table}
                <details class="rounded-md border border-line p-3" open>
                  <summary class="cursor-pointer text-sm font-semibold">{table.name} <span class="font-normal text-slate-500">({table.row_count})</span></summary>
                  <div class="mt-2 space-y-1">
                    {#each table.columns as column}
                      <div class="flex items-center justify-between gap-2 text-xs">
                        <span class="truncate">{column.name}{column.primary_key ? ' *' : ''}</span>
                        <span class="shrink-0 text-slate-500">{column.type || 'TEXT'}</span>
                      </div>
                    {/each}
                  </div>
                </details>
              {/each}
            </div>
          </aside>

          <section class="min-w-0 space-y-4">
            <div class="rounded-lg border border-line bg-white p-4 shadow-soft">
              <label class="text-sm font-semibold" for="question">Ask a question</label>
              <textarea id="question" class="focus-ring mt-2 h-24 w-full resize-y rounded-md border border-line px-3 py-2 text-sm disabled:opacity-50" bind:value={question} disabled={generating}></textarea>

              {#if generating}
                <div class="mt-3 flex min-h-[40px] items-center gap-3 rounded-md border border-accent/20 bg-accent/5 px-4 py-2.5">
                  <div class="h-4 w-4 shrink-0 animate-spin rounded-full border-2 border-accent border-t-transparent"></div>
                  {#key phraseIdx}
                    <span class="text-sm font-medium text-accent" in:fade={{ duration: 300 }}>
                      {PHRASES[phraseIdx]}
                    </span>
                  {/key}
                </div>
              {:else}
                <div class="mt-3 flex justify-end gap-2">
                  <button class="focus-ring flex items-center gap-2 rounded-md border border-line bg-white px-4 py-2 text-sm font-medium disabled:opacity-60" disabled={busy} on:click={generate}><Bot size={16} /> Generate SQL</button>
                  <button class="focus-ring flex items-center gap-2 rounded-md bg-accent px-4 py-2 text-sm font-medium text-white disabled:opacity-60" disabled={busy} on:click={generateAndRun}><Zap size={16} /> Generate &amp; Run</button>
                </div>
              {/if}
            </div>

            <div class="rounded-lg border border-line bg-white p-4 shadow-soft">
              <div class="mb-2 flex items-center justify-between">
                <label class="text-sm font-semibold" for="sql-editor">Generated SQL</label>
                <span class="text-xs text-slate-500">confidence {(confidence * 100).toFixed(0)}%</span>
              </div>
              <textarea id="sql-editor" class="focus-ring h-40 w-full resize-y rounded-md border border-line bg-slate-950 px-3 py-2 font-mono text-sm text-slate-100" bind:value={sql}></textarea>
              <div class="mt-3 flex justify-end">
                <button class="focus-ring flex items-center gap-2 rounded-md bg-ink px-4 py-2 text-sm font-medium text-white disabled:opacity-60" disabled={busy || !sql} on:click={execute}><Play size={16} /> Execute</button>
              </div>
              {#if explanation}<p class="mt-3 rounded-md bg-panel px-3 py-2 text-sm text-slate-600">{explanation}</p>{/if}
            </div>

            {#if result}
              <div class="rounded-lg border border-line bg-white shadow-soft" style="resize: vertical; overflow: hidden; min-height: 200px; display: flex; flex-direction: column;">
                <div class="flex shrink-0 items-center justify-between border-b border-line px-4 py-3 text-sm">
                  <strong>Results</strong>
                  <span class="text-slate-500">{result.row_count} rows · {result.execution_ms} ms</span>
                </div>
                <div class="min-h-0 flex-1 overflow-auto">
                  <table class="min-w-full text-left text-sm">
                    <thead class="sticky top-0 bg-panel">
                      <tr>{#each result.columns as column}<th class="border-b border-line px-3 py-2 font-semibold">{column}</th>{/each}</tr>
                    </thead>
                    <tbody>
                      {#each result.rows as row}
                        <tr class="odd:bg-white even:bg-slate-50">{#each row as cell}<td class="border-b border-line px-3 py-2">{cell === null ? 'NULL' : String(cell)}</td>{/each}</tr>
                      {/each}
                    </tbody>
                  </table>
                </div>
              </div>
            {/if}
          </section>

          <aside class="max-h-[calc(100vh-10rem)] self-start overflow-y-auto rounded-lg border border-line bg-white p-4 shadow-soft">
            <div class="mb-3 flex items-center justify-between">
              <h2 class="font-semibold">History</h2>
              {#if history.length > 0}
                <button class="focus-ring grid h-8 w-8 place-items-center rounded-md border border-line text-slate-500 hover:text-red-600" title="Clear history" aria-label="Clear history" on:click={clearHistory}><Trash2 size={14} /></button>
              {/if}
            </div>
            <div class="space-y-2">
              {#each history as item}
                <button class="focus-ring w-full rounded-md border border-line p-3 text-left text-sm hover:bg-panel" on:click={() => useHistory(item)}>
                  <div class="mb-1 flex items-center justify-between gap-2">
                    <span class="truncate font-medium">{item.question || item.status}</span>
                    <ChevronRight size={14} />
                  </div>
                  <div class="truncate font-mono text-xs text-slate-500">{item.executed_sql || item.generated_sql || item.error_message}</div>
                </button>
              {/each}
            </div>
          </aside>
        </div>
      {/if}
    {/if}
  </main>
</AppShell>
