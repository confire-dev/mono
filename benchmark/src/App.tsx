import { Routes, Route, Link, useLocation } from 'react-router-dom'
import SessionA from './pages/SessionA'
import SessionB from './pages/SessionB'

export default function App() {
  const { pathname } = useLocation()

  return (
    <div style={{ minHeight: '100vh' }}>
      <nav style={{
        display: 'flex', alignItems: 'center', gap: '1rem',
        padding: '1rem 2rem', borderBottom: '1px solid var(--border)',
        position: 'sticky', top: 0, background: 'var(--bg)', zIndex: 10,
      }}>
        <span style={{ fontFamily: 'monospace', fontSize: '0.85rem', color: 'var(--text-muted)', marginRight: '1rem' }}>
          confire/benchmark
        </span>
        <Link to="/session-a" style={{
          padding: '0.35rem 0.9rem', borderRadius: '6px', fontSize: '0.875rem',
          background: pathname === '/session-a' ? 'var(--green)' : 'var(--bg-card)',
          color: pathname === '/session-a' ? '#000' : 'var(--text-muted)',
          border: '1px solid var(--border)',
        }}>
          Session A · Confire ON
        </Link>
        <Link to="/session-b" style={{
          padding: '0.35rem 0.9rem', borderRadius: '6px', fontSize: '0.875rem',
          background: pathname === '/session-b' ? 'var(--red)' : 'var(--bg-card)',
          color: pathname === '/session-b' ? '#fff' : 'var(--text-muted)',
          border: '1px solid var(--border)',
        }}>
          Session B · Confire OFF
        </Link>
      </nav>

      <Routes>
        <Route path="/" element={<Home />} />
        <Route path="/session-a" element={<SessionA />} />
        <Route path="/session-b" element={<SessionB />} />
      </Routes>
    </div>
  )
}

function Home() {
  return (
    <div style={{ padding: '4rem 2rem', maxWidth: '640px', margin: '0 auto' }}>
      <p style={{ fontFamily: 'monospace', color: 'var(--cyan)', marginBottom: '0.5rem', fontSize: '0.875rem' }}>
        confire benchmark
      </p>
      <h1 style={{ fontSize: '2rem', fontWeight: 700, marginBottom: '1rem' }}>
        Same task. Same design. Same repo.
      </h1>
      <p style={{ color: 'var(--text-muted)', marginBottom: '2rem', lineHeight: 1.7 }}>
        Two Claude sessions implement the same component from Figma into this app.
        Session A runs with Confire active. Session B runs without.
        Everything else is identical.
      </p>
      <div style={{ display: 'flex', gap: '1rem' }}>
        <Link to="/session-a" style={{
          padding: '0.6rem 1.25rem', borderRadius: '8px',
          background: 'var(--green)', color: '#000', fontWeight: 600, fontSize: '0.9rem',
        }}>
          Session A → Confire ON
        </Link>
        <Link to="/session-b" style={{
          padding: '0.6rem 1.25rem', borderRadius: '8px',
          background: 'var(--bg-card)', color: 'var(--text-muted)',
          border: '1px solid var(--border)', fontSize: '0.9rem',
        }}>
          Session B → Confire OFF
        </Link>
      </div>
    </div>
  )
}
