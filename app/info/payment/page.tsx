import { Container } from '@/components/Container'
import { api } from '@/lib/api'

export default async function PaymentInfoPage() {
  let page = null
  let error: string | null = null

  try {
    page = await api.getPage('payment')
  } catch (e) {
    error = e instanceof Error ? e.message : 'Ошибка загрузки страницы'
    console.error('Failed to load page:', e)
  }

  return (
    <section className="section">
      <Container size="wide">
        <div className="kicker">информация</div>
        <h1 className="h2">{page?.title || 'Оплата'}</h1>
        {error ? (
          <div style={{ color: 'var(--muted)', padding: '2rem 0' }}>{error}</div>
        ) : (
          <div
            className="lead"
            style={{ marginTop: '1rem', whiteSpace: 'pre-wrap' }}
            dangerouslySetInnerHTML={{ __html: page?.content || 'Информация об оплате скоро появится.' }}
          />
        )}
      </Container>
    </section>
  )
}
