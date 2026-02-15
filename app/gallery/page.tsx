import { Container } from '@/components/Container'

const TABS = ['Интро', 'Интро 2', 'Интро 3', 'ТАТУ', 'Токио', 'Прохор', 'Назар'] as const

export default function GalleryPage() {
  return (
    <section className="section">
      <Container size="wide">
        <div className="kicker">галерея</div>
        <h1 className="h2">Смотреть</h1>
        <p className="lead">
          TODO: табы и грид картинок. Эндпоинт: <code>/api/gallery?category=intro</code> (категорию нормализуем в slug).
        </p>

        <div style={{ marginTop: '1.25rem', display: 'flex', flexWrap: 'wrap', gap: '0.5rem' }}>
          {TABS.map((t) => (
            <span
              key={t}
              style={{
                border: '1px solid var(--line)',
                padding: '0.35rem 0.6rem',
                borderRadius: 999,
                fontSize: '0.9rem',
                color: 'var(--muted)'
              }}
            >
              {t}
            </span>
          ))}
        </div>
      </Container>
    </section>
  )
}

