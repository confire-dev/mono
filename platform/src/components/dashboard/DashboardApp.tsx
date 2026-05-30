"use client"

import { useState, useEffect } from "react"
import { SidebarProvider, SidebarInset } from "@/components/ui/sidebar"
import { AppSidebar } from "@/components/app-sidebar"
import { SiteHeader } from "@/components/site-header"
import { SectionCards } from "@/components/section-cards"
import { ChartAreaInteractive } from "@/components/chart-area-interactive"
import { DataTable } from "@/components/data-table"
import { useAuth } from "@/hooks/use-auth"
import data from "@/app/dashboard/data.json"

const PAGE_TITLES: Record<string, string> = {
  overview: "Overview",
  sessions: "CLI Sessions",
  billing:  "Billing",
  settings: "Settings",
}

function getInitialPage() {
  if (typeof window === "undefined") return "overview"
  const hash = window.location.hash.slice(1)
  return hash in PAGE_TITLES ? hash : "overview"
}

export function DashboardApp() {
  const { user, loading, signOut } = useAuth()
  const [page, setPage] = useState(getInitialPage)

  useEffect(() => {
    const onHashChange = () => {
      const hash = window.location.hash.slice(1)
      if (hash in PAGE_TITLES) setPage(hash)
    }
    window.addEventListener("hashchange", onHashChange)
    return () => window.removeEventListener("hashchange", onHashChange)
  }, [])

  const navigate = (p: string) => {
    window.location.hash = p
    setPage(p)
  }

  if (loading) {
    return (
      <div className="flex h-screen items-center justify-center">
        <p className="text-sm text-muted-foreground">Loading…</p>
      </div>
    )
  }

  if (!user) {
    window.location.href = "/login"
    return null
  }

  return (
    <SidebarProvider>
      <AppSidebar
        variant="inset"
        user={{
          name:  user.user_metadata?.["name"] ?? "",
          email: user.email ?? "",
        }}
        currentPage={page}
        onNavigate={navigate}
        onSignOut={signOut}
      />
      <SidebarInset>
        <SiteHeader title={PAGE_TITLES[page]} />
        <div className="flex flex-1 flex-col">
          <div className="@container/main flex flex-1 flex-col gap-2">
            <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
              {page === "overview" && (
                <>
                  <SectionCards />
                  <div className="px-4 lg:px-6">
                    <ChartAreaInteractive />
                  </div>
                  <DataTable data={data} />
                </>
              )}
              {page === "sessions" && (
                <div className="px-4 lg:px-6">
                  <p className="text-sm text-muted-foreground">Sessions — coming soon.</p>
                </div>
              )}
              {page === "billing" && (
                <div className="px-4 lg:px-6">
                  <p className="text-sm text-muted-foreground">Billing — coming soon.</p>
                </div>
              )}
              {page === "settings" && (
                <div className="px-4 lg:px-6">
                  <p className="text-sm text-muted-foreground">Settings — coming soon.</p>
                </div>
              )}
            </div>
          </div>
        </div>
      </SidebarInset>
    </SidebarProvider>
  )
}
