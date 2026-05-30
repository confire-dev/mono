import {
  IconBolt,
  IconCreditCard,
  IconDashboard,
  IconHelp,
  IconSettings,
  IconTerminal2,
} from '@tabler/icons-react'
import { NavUser } from '@/components/nav-user'
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarGroup,
  SidebarGroupContent,
} from '@/components/ui/sidebar'

const NAV = [
  { title: 'Overview',  url: '/dashboard',          icon: IconDashboard   },
  { title: 'Sessions',  url: '/dashboard/sessions',  icon: IconTerminal2   },
  { title: 'Billing',   url: '/dashboard/billing',   icon: IconCreditCard  },
  { title: 'Settings',  url: '/dashboard/settings',  icon: IconSettings    },
]

const NAV_SECONDARY = [
  { title: 'Help',  url: 'https://confire.dev/docs', icon: IconHelp },
]

interface Props extends React.ComponentProps<typeof Sidebar> {
  user: { name?: string; email: string; avatar?: string }
  currentPath: string
}

export function ConFireSidebar({ user, currentPath, ...props }: Props) {
  return (
    <Sidebar collapsible="offcanvas" {...props}>
      <SidebarHeader>
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton asChild className="data-[slot=sidebar-menu-button]:p-1.5!">
              <a href="/">
                <IconBolt className="size-5!" />
                <span className="text-base font-semibold">Confire</span>
              </a>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarHeader>

      <SidebarContent>
        <SidebarGroup>
          <SidebarGroupContent>
            <SidebarMenu>
              {NAV.map(item => (
                <SidebarMenuItem key={item.title}>
                  <SidebarMenuButton
                    asChild
                    isActive={currentPath === item.url || currentPath.startsWith(item.url + '/')}
                  >
                    <a href={item.url}>
                      <item.icon />
                      <span>{item.title}</span>
                    </a>
                  </SidebarMenuButton>
                </SidebarMenuItem>
              ))}
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>

        <SidebarGroup className="mt-auto">
          <SidebarGroupContent>
            <SidebarMenu>
              {NAV_SECONDARY.map(item => (
                <SidebarMenuItem key={item.title}>
                  <SidebarMenuButton asChild>
                    <a href={item.url} target="_blank" rel="noreferrer">
                      <item.icon />
                      <span>{item.title}</span>
                    </a>
                  </SidebarMenuButton>
                </SidebarMenuItem>
              ))}
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
      </SidebarContent>

      <SidebarFooter>
        <NavUser user={{ name: user.name ?? user.email, email: user.email, avatar: user.avatar ?? '' }} />
      </SidebarFooter>
    </Sidebar>
  )
}
