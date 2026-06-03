import { useMemo, useState } from 'react'
import { Button, Field, Input, Text } from '@cloudflare/kumo'
import { createBrowserClient } from '@/lib/supabase'
import { cn } from '@/lib/utils'
import { AppShell } from '@/components/app/AppShell'

type Step = 'idle' | 'email-sent' | 'submitting'

interface Props {
  className?: string
  planIntent?: string
  next?: string
}

export function AuthForm({ className, planIntent = 'free', next = '' }: Props) {
  const supabase = useMemo(() => createBrowserClient(), [])
  const [step, setStep] = useState<Step>('idle')
  const [email, setEmail] = useState('')
  const [otp, setOtp] = useState('')
  const [error, setError] = useState('')

  const isFree = planIntent === 'free'

  const cliParams = useMemo(() => {
    if (!next) return null
    try {
      const nextUrl = new URL(next, 'http://x')
      if (!nextUrl.pathname.startsWith('/cli/login')) return null
      return {
        deviceName: nextUrl.searchParams.get('device_name') ?? '',
        cliVersion: nextUrl.searchParams.get('cli_version') ?? '',
      }
    } catch { return null }
  }, [next])

  const isCli = !!cliParams

  const callbackUrl = typeof window !== 'undefined'
    ? `${window.location.origin}/auth/callback?plan=${encodeURIComponent(planIntent)}${next ? `&next=${encodeURIComponent(next)}` : ''}`
    : `/auth/callback?plan=${encodeURIComponent(planIntent)}${next ? `&next=${encodeURIComponent(next)}` : ''}`

  const isLoading = step === 'submitting'

  async function signInWithProvider(provider: 'github' | 'google') {
    setError('')
    const { error: authError } = await supabase.auth.signInWithOAuth({
      provider,
      options: { redirectTo: callbackUrl },
    })
    if (authError) setError(authError.message)
  }

  async function sendEmail(e: React.FormEvent) {
    e.preventDefault()
    if (!email) return
    setStep('submitting')
    setError('')
    const { error: authError } = await supabase.auth.signInWithOtp({
      email,
      options: { shouldCreateUser: true, emailRedirectTo: callbackUrl },
    })
    if (authError) { setError(authError.message); setStep('idle') }
    else setStep('email-sent')
  }

  async function verifyOtp(e: React.FormEvent) {
    e.preventDefault()
    if (!otp || otp.length < 6) return
    setStep('submitting')
    setError('')
    const { error: authError } = await supabase.auth.verifyOtp({ email, token: otp, type: 'email' })
    if (authError) { setError(authError.message); setStep('email-sent') }
    else window.location.href = callbackUrl
  }

  const title = step === 'email-sent'
    ? 'Check your email'
    : isCli
      ? 'Sign in to Authorize CLI'
      : !isFree
        ? 'Continue to checkout'
        : 'Sign in to Confire'

  return (
    <AppShell>
      <div className={cn('flex flex-col gap-5', className)}>
        <Text variant="heading2" as="h1" DANGEROUS_className="text-center">{title}</Text>

        {step === 'email-sent' ? (
          <form onSubmit={verifyOtp} className="flex flex-col gap-4">
            <Text variant="secondary" size="sm" as="p" DANGEROUS_className="text-center">
              We sent a 6-digit code to <strong style={{ color: '#fff' }}>{email}</strong>
            </Text>
            <Field label="One-time code">
              <Input
                type="text"
                inputMode="numeric"
                pattern="[0-9]*"
                maxLength={6}
                placeholder="123456"
                value={otp}
                onChange={e => setOtp(e.target.value.replace(/\D/g, ''))}
                autoFocus
                required
              />
            </Field>
            <Button type="submit" variant="primary" className="w-full" loading={isLoading}>
              Verify code
            </Button>
            <button
              type="button"
              onClick={() => { setStep('idle'); setOtp('') }}
              className="text-center text-sm text-kumo-subtle underline-offset-4 hover:underline"
            >
              Use a different email
            </button>
          </form>
        ) : (
          <>
            {/* GitHub — full width */}
            <Button
              variant="outline"
              className="w-full"
              disabled={isLoading}
              onClick={() => signInWithProvider('github')}
              type="button"
            >
              <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" width="16" height="16" style={{ marginRight: 8, flexShrink: 0 }}>
                <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z" fill="currentColor" />
              </svg>
              Continue with GitHub
            </Button>

            {/* Google — full width */}
            <Button
              variant="outline"
              className="w-full"
              disabled={isLoading}
              onClick={() => signInWithProvider('google')}
              type="button"
            >
              <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" width="16" height="16" style={{ marginRight: 8, flexShrink: 0 }}>
                <path d="M12.48 10.92v3.28h7.84c-.24 1.84-.853 3.187-1.787 4.133-1.147 1.147-2.933 2.4-6.053 2.4-4.827 0-8.6-3.893-8.6-8.72s3.773-8.72 8.6-8.72c2.6 0 4.507 1.027 5.907 2.347l2.307-2.307C18.747 1.44 16.133 0 12.48 0 5.867 0 .307 5.387.307 12s5.56 12 12.173 12c3.573 0 6.267-1.173 8.373-3.36 2.16-2.16 2.84-5.213 2.84-7.667 0-.76-.053-1.467-.173-2.053H12.48z" fill="currentColor" />
              </svg>
              Continue with Google
            </Button>

            {/* or divider */}
            <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
              <div style={{ flex: 1, height: 1, background: 'var(--kumo-hairline, rgba(255,255,255,0.1))' }} />
              <Text variant="secondary" size="sm" as="span">or</Text>
              <div style={{ flex: 1, height: 1, background: 'var(--kumo-hairline, rgba(255,255,255,0.1))' }} />
            </div>

            {/* email form */}
            <form onSubmit={sendEmail} className="flex flex-col gap-3">
              <Field label="Email">
                <Input
                  type="email"
                  placeholder="you@example.com"
                  value={email}
                  onChange={e => setEmail(e.target.value)}
                  required
                  autoComplete="email"
                />
              </Field>
              <Button type="submit" variant="primary" className="w-full" loading={isLoading}>
                Continue with email
              </Button>
            </form>
          </>
        )}

        {error && (
          <Text variant="error" size="sm" as="p" DANGEROUS_className="text-center">
            {error}
          </Text>
        )}

        <Text variant="secondary" size="xs" as="p" DANGEROUS_className="text-balance text-center [&_a]:underline [&_a]:underline-offset-4 hover:[&_a]:text-kumo-link">
          By continuing, you agree to our{' '}
          <a href="/terms">Terms of Service</a> and{' '}
          <a href="/privacy">Privacy Policy</a>.
        </Text>
      </div>
    </AppShell>
  )
}
