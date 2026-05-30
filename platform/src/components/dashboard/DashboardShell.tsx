import { SidebarProvider, SidebarInset, SidebarTrigger } from '@/components/ui/sidebar'
import { Separator } from '@/components/ui/separator'
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbList,
  BreadcrumbPage,
} from '@/components/ui/breadcrumb'
import { ConFireSidebar } from './ConFireSidebar'

interface Props {
  children: React.ReactNode
  title: string
  currentPath: string
  user: { email: string; name?: string; avatar?: string }
}

export function DashboardShell({ children, title, currentPath, user }: Props) {
  return (
    <SidebarProvider>
      <ConFireSidebar currentPath={currentPath} user={user} />
      <SidebarInset>
        {/* Top bar: hamburger + breadcrumb */}
        <header className="flex h-12 shrink-0 items-center gap-2 border-b px-4">
          <SidebarTrigger className="-ml-1" />
          <Separator orientation="vertical" className="mr-2 h-4" />
          <Breadcrumb>
            <BreadcrumbList>
              <BreadcrumbItem>
                <BreadcrumbPage>{title}</BreadcrumbPage>
              </BreadcrumbItem>
            </BreadcrumbList>
          </Breadcrumb>
        </header>

        {/* Page content */}
        <div className="flex flex-1 flex-col gap-4 p-4 lg:p-6">
          {children}
        </div>
      </SidebarInset>
    </SidebarProvider>
  )
}
