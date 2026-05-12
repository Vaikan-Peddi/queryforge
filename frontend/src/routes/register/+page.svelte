<script lang="ts">
  import { goto } from '$app/navigation';
  import { UserPlus } from 'lucide-svelte';
  import { api, saveSession } from '$lib/api';

  let name = '';
  let email = '';
  let password = '';
  let error = '';
  let loading = false;

  async function submit() {
    loading = true;
    error = '';
    try {
      saveSession(await api.register(name, email, password));
      goto('/dashboard');
    } catch (err) {
      error = err instanceof Error ? err.message : 'Registration failed';
    } finally {
      loading = false;
    }
  }
</script>

<main class="grid min-h-screen place-items-center px-4">
  <section class="w-full max-w-md rounded-lg border border-line bg-white p-8 shadow-soft">
    <div class="mb-8 flex items-center gap-3">
      <img class="h-11 w-11 rounded-md" src="/brand/queryforge-logo.svg" alt="QueryForge logo" />
      <div>
        <h1 class="text-2xl font-semibold">Create Account</h1>
        <p class="text-sm text-slate-500">Start a QueryForge workspace</p>
      </div>
    </div>
    <form class="space-y-4" on:submit|preventDefault={submit}>
      <label class="block text-sm font-medium">Name<input class="focus-ring mt-1 w-full rounded-md border border-line px-3 py-2" bind:value={name} required /></label>
      <label class="block text-sm font-medium">Email<input class="focus-ring mt-1 w-full rounded-md border border-line px-3 py-2" bind:value={email} type="email" required /></label>
      <label class="block text-sm font-medium">Password<input class="focus-ring mt-1 w-full rounded-md border border-line px-3 py-2" bind:value={password} type="password" minlength="8" required /></label>
      {#if error}<p class="rounded-md bg-red-50 px-3 py-2 text-sm text-red-700">{error}</p>{/if}
      <button class="focus-ring flex w-full items-center justify-center gap-2 rounded-md bg-ink px-4 py-2.5 font-medium text-white disabled:opacity-60" disabled={loading}>
        <UserPlus size={18} /> {loading ? 'Creating' : 'Create account'}
      </button>
    </form>
    <p class="mt-6 text-center text-sm text-slate-500">Already registered? <a class="font-medium text-accent" href="/login">Sign in</a></p>
  </section>
</main>
