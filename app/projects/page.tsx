import Link from 'next/link'
import { Container } from '@/components/Container'

const MOCK_PROJECTS = [
  { slug: 'den-pobedy', title: 'День победы' },
  { slug: 'bikkembergs', title: 'Bikkembergs' }
]

export default function ProjectsPage() {
  return (
    <section className="section">
      <Container size="wide">
        <div className="kicker">проекты</div>
        <h1 className="h2">Смотреть</h1>
        <p className="lead">
          TODO: список проектов из <code>/api/projects</code>. Пока мок.
        </p>

        <div style={{ marginTop: '1.25rem', display: 'grid', gap: '0.75rem' }}>
          {MOCK_PROJECTS.map((p) => (
            <Link
              key={p.slug}
              href={`/projects/${p.slug}`}
              style={{
                border: '1px solid var(--line)',
                borderRadius: 'var(--radius)',
                padding: '1rem',
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'baseline'
              }}
            >
              <span>{p.title}</span>
              <span style={{ color: 'var(--muted)', fontFamily: 'var(--mono)', letterSpacing: '0.12em' }}>
                open
              </span>
            </Link>
          ))}
        </div>
      </Container>
    </section>
  )
}

