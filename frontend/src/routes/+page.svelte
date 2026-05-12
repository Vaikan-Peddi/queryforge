<script lang="ts">
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { getSession } from '$lib/api';

  onMount(() => {
    goto(getSession() ? '/dashboard' : '/login');
  });
</script>

<main class="grid min-h-screen place-items-center">
  <div class="h-10 w-10 animate-spin rounded-full border-2 border-accent border-t-transparent"></div>
</main>
