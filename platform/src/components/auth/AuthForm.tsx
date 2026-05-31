import { useMemo, useState } from "react"
import { Zap } from "lucide-react"
import { createBrowserClient } from "@/lib/supabase"
import { cn } from "@/lib/utils"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"

type Step = "idle" | "email-sent" | "submitting"

interface Props {
  className?: string
  planIntent?: string
  next?: string
}

export function AuthForm({ className, planIntent = "free", next = "" }: Props) {
  const supabase = useMemo(() => createBrowserClient(), [])
  const [step,  setStep]  = useState<Step>("idle")
  const [email, setEmail] = useState("")
  const [otp,   setOtp]   = useState("")
  const [error, setError] = useState("")

  const isFree = planIntent === "free"

  // Auth callback carries plan intent + next so the server can route correctly
  // after onboarding (e.g. back to the CLI authorize page).
  const callbackUrl = typeof window !== "undefined"
    ? `${window.location.origin}/auth/callback?plan=${encodeURIComponent(planIntent)}${next ? `&next=${encodeURIComponent(next)}` : ''}`
    : `/auth/callback?plan=${encodeURIComponent(planIntent)}${next ? `&next=${encodeURIComponent(next)}` : ''}`

  const isLoading = step === "submitting"

  async function signInWithProvider(provider: "github" | "google") {
    setError("")
    const { error } = await supabase.auth.signInWithOAuth({
      provider,
      options: { redirectTo: callbackUrl },
    })
    if (error) setError(error.message)
  }

  async function sendEmail(e: React.FormEvent) {
    e.preventDefault()
    if (!email) return
    setStep("submitting")
    setError("")
    const { error } = await supabase.auth.signInWithOtp({
      email,
      options: { shouldCreateUser: true, emailRedirectTo: callbackUrl },
    })
    if (error) { setError(error.message); setStep("idle") }
    else setStep("email-sent")
  }

  async function verifyOtp(e: React.FormEvent) {
    e.preventDefault()
    if (!otp || otp.length < 6) return
    setStep("submitting")
    setError("")
    const { error } = await supabase.auth.verifyOtp({ email, token: otp, type: "email" })
    if (error) { setError(error.message); setStep("email-sent") }
    // Navigate to the callback so the server handles plan-based routing.
    else window.location.href = callbackUrl
  }

  const title = step === "email-sent"
    ? "Check your email"
    : isFree
      ? "Start free with Confire"
      : "Continue to checkout"

  const subtitle = step === "email-sent"
    ? `We sent a code to ${email}`
    : isFree
      ? "Install Confire for Claude Code and optimize your first tool output."
      : "Log in first, then we'll send you to secure checkout."

  return (
    <div className={cn("flex flex-col gap-6", className)}>
      {/* Brand */}
      <div className="flex flex-col items-center gap-2">
        <a href="/" className="flex flex-col items-center gap-2 font-medium">
          <div className="flex size-8 items-center justify-center rounded-md bg-primary text-primary-foreground">
            <Zap className="size-5" />
          </div>
          <span className="sr-only">Confire</span>
        </a>
        <h1 className="text-xl font-bold">{title}</h1>
        <p className="text-balance text-center text-sm text-muted-foreground">{subtitle}</p>
      </div>

      {/* OTP verify */}
      {step === "email-sent" && (
        <form onSubmit={verifyOtp} className="flex flex-col gap-4">
          <div className="grid gap-2">
            <Label htmlFor="otp">One-time code</Label>
            <Input
              id="otp"
              type="text"
              inputMode="numeric"
              pattern="[0-9]*"
              maxLength={6}
              placeholder="123456"
              value={otp}
              onChange={e => setOtp(e.target.value.replace(/\D/g, ""))}
              autoFocus
              required
            />
          </div>
          <Button type="submit" className="w-full" disabled={isLoading}>
            {isLoading ? "Verifying…" : "Verify code"}
          </Button>
          <button
            type="button"
            onClick={() => { setStep("idle"); setOtp("") }}
            className="text-center text-sm text-muted-foreground underline-offset-4 hover:underline"
          >
            Use a different email
          </button>
        </form>
      )}

      {/* Email + socials */}
      {step !== "email-sent" && (
        <>
          <form onSubmit={sendEmail} className="flex flex-col gap-4">
            <div className="grid gap-2">
              <Label htmlFor="email">Email</Label>
              <Input
                id="email"
                type="email"
                placeholder="you@example.com"
                value={email}
                onChange={e => setEmail(e.target.value)}
                required
                autoComplete="email"
              />
            </div>
            <Button type="submit" className="w-full" disabled={isLoading}>
              {isLoading ? "Sending…" : "Continue with email"}
            </Button>
          </form>

          <div className="relative text-center text-sm after:absolute after:inset-0 after:top-1/2 after:z-0 after:flex after:items-center after:border-t after:border-border">
            <span className="relative z-10 bg-background px-2 text-muted-foreground">Or</span>
          </div>

          <div className="grid gap-4 sm:grid-cols-2">
            <Button variant="outline" className="w-full" disabled={isLoading} onClick={() => signInWithProvider("github")} type="button">
              <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" className="size-4">
                <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z" fill="currentColor"/>
              </svg>
              Continue with GitHub
            </Button>
            <Button variant="outline" className="w-full" disabled={isLoading} onClick={() => signInWithProvider("google")} type="button">
              <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" className="size-4">
                <path d="M12.48 10.92v3.28h7.84c-.24 1.84-.853 3.187-1.787 4.133-1.147 1.147-2.933 2.4-6.053 2.4-4.827 0-8.6-3.893-8.6-8.72s3.773-8.72 8.6-8.72c2.6 0 4.507 1.027 5.907 2.347l2.307-2.307C18.747 1.44 16.133 0 12.48 0 5.867 0 .307 5.387.307 12s5.56 12 12.173 12c3.573 0 6.267-1.173 8.373-3.36 2.16-2.16 2.84-5.213 2.84-7.667 0-.76-.053-1.467-.173-2.053H12.48z" fill="currentColor"/>
              </svg>
              Continue with Google
            </Button>
          </div>
        </>
      )}

      {error && <p className="text-center text-sm text-destructive">{error}</p>}

      <p className="text-balance text-center text-xs text-muted-foreground [&_a]:underline [&_a]:underline-offset-4 hover:[&_a]:text-primary">
        By clicking continue, you agree to our{" "}
        <a href="/terms">Terms of Service</a> and{" "}
        <a href="/privacy">Privacy Policy</a>.
      </p>
    </div>
  )
}
