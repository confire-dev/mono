import * as React from 'react'
import { cn } from '@/lib/utils'
import { WithCorners } from './corner-squares'
import { CodeHighlighted, ShikiProvider } from '@cloudflare/kumo/code'

export function CodeBlock({
  code,
  lang = 'bash',
  fileTabs,
  activeFile = 0,
  onFileChange,
  className,
}: {
  code: string
  lang?: string
  fileTabs?: string[]
  activeFile?: number
  onFileChange?: (index: number) => void
  className?: string
}) {
  return (
    <ShikiProvider languages={['bash', 'typescript', 'javascript', 'json', 'sh']}>
      <div
        className={cn(
          'flex h-full flex-col overflow-hidden bg-confire-code font-mono',
          className,
        )}
      >
        {fileTabs && fileTabs.length > 0 && (
          <div className="flex shrink-0 gap-0 border-b border-confire-border-subtle px-4">
            {fileTabs.map((tab, i) => (
              <button
                key={tab}
                type="button"
                onClick={() => onFileChange?.(i)}
                className={cn(
                  'cursor-pointer border-none bg-transparent px-4 py-2.5 font-sans text-xs transition-colors',
                  i === activeFile
                    ? 'border-b-2 border-confire-accent font-semibold text-confire-text'
                    : 'border-b-2 border-transparent text-confire-code-tab-inactive hover:text-confire-dim',
                )}
              >
                {tab}
              </button>
            ))}
          </div>
        )}

        {/* force-dark: override shiki dual-theme to always use dark colors */}
        <div
          className="relative flex-1 overflow-y-auto [&_code_span]:![color:var(--shiki-dark)] [&_pre]:!bg-transparent"
          data-theme="dark"
        >
          <CodeHighlighted
            code={code}
            language={lang}
            showCopyButton
            className="!rounded-none !border-0 !bg-transparent"
          />
        </div>
      </div>
    </ShikiProvider>
  )
}

export interface TabItem {
  label: string
  icon?: React.ReactNode
  content: React.ReactNode
}

export function HorizontalTabs({
  tabs,
  defaultTab = 0,
}: {
  tabs: TabItem[]
  defaultTab?: number
}) {
  const [active, setActive] = React.useState(defaultTab)

  return (
    <div>
      <div className="mb-10 flex justify-center">
        <div className="inline-flex gap-0.5 rounded-full border border-confire-border bg-confire-card p-1">
          {tabs.map((tab, i) => (
            <button
              key={tab.label}
              type="button"
              onClick={() => setActive(i)}
              className={cn(
                'flex cursor-pointer items-center gap-1.5 rounded-full border-none px-[22px] py-2 font-sans text-sm font-semibold transition-all duration-150',
                i === active
                  ? 'bg-confire-accent text-confire-white'
                  : 'bg-transparent text-confire-secondary hover:text-confire-dim',
              )}
            >
              {tab.icon && <span className="text-[15px]">{tab.icon}</span>}
              {tab.label}
            </button>
          ))}
        </div>
      </div>
      <div className="transition-opacity duration-200">{tabs[active]?.content}</div>
    </div>
  )
}

export interface VerticalTabItem {
  title: string
  description: string
  lang?: string
  code?: string
  files?: { name: string; lang?: string; code: string }[]
}

export function VerticalTabsCode({
  tabs,
  defaultTab = 0,
}: {
  tabs: VerticalTabItem[]
  defaultTab?: number
}) {
  const [active, setActive] = React.useState(defaultTab)
  const [activeFile, setActiveFile] = React.useState(0)
  const current = tabs[active]
  const fileTabs = current?.files?.map((f) => f.name)
  const code = current?.files?.[activeFile]?.code ?? current?.code ?? ''
  const lang = current?.files?.[activeFile]?.lang ?? current?.lang ?? 'bash'

  return (
    <WithCorners cols={2} rows={1}>
      <div className="grid min-h-[420px] grid-cols-1 border border-confire-border md:grid-cols-2">
        <div className="overflow-hidden border-b border-confire-border md:min-h-[280px] md:border-b-0 md:border-r">
          <CodeBlock
            code={code}
            lang={lang}
            fileTabs={fileTabs}
            activeFile={activeFile}
            onFileChange={setActiveFile}
          />
        </div>
        <div className="flex flex-col">
          {tabs.map((tab, i) => (
            <button
              key={tab.title}
              type="button"
              onClick={() => {
                setActive(i)
                setActiveFile(0)
              }}
              className={cn(
                'cursor-pointer border-none bg-transparent text-left transition-all duration-150',
                'border-b border-confire-border-subtle last:border-b-0',
                i === active
                  ? 'border-l-2 border-l-confire-accent bg-confire-accent-dim'
                  : 'border-l-2 border-l-transparent hover:bg-confire-hover-subtle',
              )}
            >
              <div className="px-6 py-5">
                <div
                  className={cn(
                    'mb-1 text-[15px] font-semibold transition-colors',
                    i === active ? 'text-confire-text' : 'text-confire-inactive',
                  )}
                >
                  {tab.title}
                </div>
                <p className="m-0 text-[13px] leading-snug text-confire-subtle">{tab.description}</p>
              </div>
            </button>
          ))}
        </div>
      </div>
    </WithCorners>
  )
}

export function VerticalListTabs({
  tabs,
  defaultTab = 0,
}: {
  tabs: { title: string; content: React.ReactNode }[]
  defaultTab?: number
}) {
  const [active, setActive] = React.useState(defaultTab)

  return (
    <WithCorners cols={1} rows={1}>
      <div className="border border-confire-border">
        {tabs.map((tab, i) => (
          <div
            key={tab.title}
            className={cn(
              'border-b border-confire-border-subtle last:border-b-0',
              i === active ? 'border-l-2 border-l-confire-accent' : 'border-l-2 border-l-transparent',
            )}
          >
            <button
              type="button"
              onClick={() => setActive(active === i ? -1 : i)}
              className="flex w-full cursor-pointer items-center justify-between border-none bg-transparent px-7 py-5 text-left"
            >
              <span
                className={cn(
                  'text-[15px] font-semibold',
                  i === active ? 'text-confire-text' : 'text-confire-tertiary',
                )}
              >
                {tab.title}
              </span>
              <svg
                width="14"
                height="14"
                viewBox="0 0 14 14"
                fill="none"
                className={cn(
                  'text-confire-subtle transition-transform duration-200',
                  i === active && 'rotate-180',
                )}
              >
                <path
                  d="M3 5l4 4 4-4"
                  stroke="currentColor"
                  strokeWidth="1.5"
                  strokeLinecap="round"
                />
              </svg>
            </button>
            {i === active && (
              <div className="px-7 pb-5 text-sm leading-relaxed text-confire-secondary">{tab.content}</div>
            )}
          </div>
        ))}
      </div>
    </WithCorners>
  )
}
