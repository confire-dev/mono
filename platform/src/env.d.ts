/// <reference types="astro/client" />

declare namespace App {
  interface Locals {
    user?: import('@supabase/supabase-js').User
    supabase?: import('@supabase/supabase-js').SupabaseClient
  }
}

interface ImportMetaEnv {
  readonly PUBLIC_SUPABASE_URL: string
  readonly PUBLIC_SUPABASE_ANON_KEY: string
  readonly SUPABASE_SERVICE_KEY: string
  readonly PUBLIC_STRIPE_PUBLISHABLE_KEY: string
  readonly STRIPE_SECRET_KEY: string
  readonly WORKER_URL: string
  readonly PUBLIC_WORKER_URL: string
}
