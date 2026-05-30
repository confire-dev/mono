"use client"

import * as React from "react"
import {
  LayoutDashboard,
  Terminal,
  CreditCard,
  Settings,
  HelpCircle,
  Search,
  Zap,
} from "lucide-react"

import { NavMain } from "@/components/nav-main"
import { NavSecondary } from "@/components/nav-secondary"
import { NavUser } from "@/components/nav-user"
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/components/ui/sidebar"

const navMain = [
  { title: "Overview", page: "overview", icon: LayoutDashboard },
  { title: "Sessions", page: "sessions", icon: Terminal },
  { title: "Billing",  page: "billing",  icon: CreditCard },
  { title: "Settings", page: "settings", icon: Settings },
]

const navSecondary = [
  { title: "Support", url: "mailto:support@confire.dev", icon: HelpCircle },
  { title: "Search",  url: "#", icon: Search },
]

interface AppSidebarProps extends React.ComponentProps<typeof Sidebar> {
  user: { name: string; email: string }
  currentPage: string
  onNavigate: (page: string) => void
  onSignOut: () => Promise<void>
}

export function AppSidebar({ user, currentPage, onNavigate, onSignOut, ...props }: AppSidebarProps) {
  return (
    <Sidebar collapsible="offcanvas" {...props}>
      <SidebarHeader>
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton
              asChild
              className="data-[slot=sidebar-menu-button]:p-1.5!"
            >
              <a href="/">
                <Zap className="size-5!" />
                <span className="text-base font-semibold">Confire</span>
              </a>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarHeader>
      <SidebarContent>
        <NavMain items={navMain} currentPage={currentPage} onNavigate={onNavigate} />
        <NavSecondary items={navSecondary} className="mt-auto" />
      </SidebarContent>
      <SidebarFooter>
        <NavUser user={user} onSignOut={onSignOut} />
      </SidebarFooter>
    </Sidebar>
  )
}
