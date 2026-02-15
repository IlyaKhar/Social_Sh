import { Container } from '@/components/Container'

export default function ShopSalePage() {
  return (
    <section className="section">
      <Container size="wide">
        <div className="kicker">магазин</div>
        <h1 className="h2">Сезонные скидки</h1>
        <p className="lead">
          TODO: фильтр <code>isOnSale=true</code>. Эндпоинт: <code>/api/products?sale=true</code>.
        </p>
      </Container>
    </section>
  )
}

