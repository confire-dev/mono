import { Button, LinkButton, Text } from '@cloudflare/kumo'
import { CheckIcon } from '@phosphor-icons/react'
import { AppShell } from '@/components/app/AppShell'

interface Props {
  isActive: boolean
}

export function BillingSuccess({ isActive }: Props) {
  return (
    <AppShell>
      <div className="flex flex-col items-center gap-6 text-center">
        <div className="flex size-12 items-center justify-center rounded-full bg-kumo-success-tint/70">
          <CheckIcon className="size-6 text-kumo-success" weight="bold" />
        </div>

        {isActive ? (
          <>
            <div>
              <Text variant="heading2" as="h1">You're all set!</Text>
              <Text variant="secondary" size="sm" as="p" DANGEROUS_className="mt-2">
                Your plan is now active.
              </Text>
            </div>
            <LinkButton href="/onboarding" variant="primary">
              Get started
            </LinkButton>
          </>
        ) : (
          <>
            <div>
              <Text variant="heading2" as="h1">Payment received</Text>
              <Text variant="secondary" size="sm" as="p" DANGEROUS_className="mt-2">
                Activating your plan — this usually takes a few seconds.
              </Text>
            </div>
            <div className="flex w-full flex-col gap-3">
              <Button variant="primary" onClick={() => window.location.reload()}>
                Refresh status
              </Button>
              <LinkButton href="/onboarding" variant="ghost" className="text-sm">
                Continue anyway
              </LinkButton>
            </div>
          </>
        )}
      </div>
    </AppShell>
  )
}

export function BillingCancelled() {
  return (
    <AppShell>
      <div className="flex flex-col items-center gap-6 text-center">
        <div>
          <Text variant="heading2" as="h1">No charge was made</Text>
          <Text variant="secondary" size="sm" as="p" DANGEROUS_className="mt-2">
            You cancelled checkout. Your account is still on the Free plan.
          </Text>
        </div>
        <div className="flex w-full flex-col gap-3">
          <LinkButton href="/onboarding" variant="primary">
            Continue on Free
          </LinkButton>
          <LinkButton href="/#pricing" variant="outline">
            Try a paid plan again
          </LinkButton>
        </div>
      </div>
    </AppShell>
  )
}
