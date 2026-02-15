import { Container } from '@/components/Container'

export default function DeliveryInfoPage() {
  return (
    <section className="section">
      <Container size="wide">
        <div className="kicker">информация</div>
        <h1 className="h2">Доставка</h1>
        <p className="lead">
          TODO: текст о доставке. Данные можно брать из эндпоинта <code>/api/pages/delivery</code> (Markdown/HTML).
        </p>
      </Container>
    </section>
  )
}

