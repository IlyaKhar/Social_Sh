import Link from 'next/link'
import { Container } from '@/components/Container'

export default function ShopAllPage() {
  return (
    <section className="section">
      <Container size="wide">
        <div className="kicker">магазин</div>
        <h1 className="h2">Смотреть все</h1>
        <p className="lead">
          TODO: тут будет грид товаров. Данные забираем из Go Fiber: <code>/api/products</code> + пагинация.
        </p>

        <div style={{ marginTop: '1rem', display: 'flex', gap: '0.75rem', flexWrap: 'wrap' }}>
          <Link href="/shop/new">Новые поступления</Link>
          <Link href="/shop/sale">Сезонные скидки</Link>
        </div>
      </Container>
    </section>
  )
}

