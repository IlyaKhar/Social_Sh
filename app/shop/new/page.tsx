import { Container } from '@/components/Container'

export default function ShopNewPage() {
  return (
    <section className="section">
      <Container size="wide">
        <div className="kicker">магазин</div>
        <h1 className="h2">Новые поступления</h1>
        <p className="lead">
          TODO: фильтр <code>isNew=true</code>. Эндпоинт: <code>/api/products?new=true</code>.
        </p>
      </Container>
    </section>
  )
}

