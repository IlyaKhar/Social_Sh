import { Container } from '@/components/Container'

export default function PaymentInfoPage() {
  return (
    <section className="section">
      <Container size="wide">
        <div className="kicker">информация</div>
        <h1 className="h2">Оплата</h1>
        <p className="lead">
          TODO: текст об оплате. Можно хранить как Markdown в Go Fiber и отдавать через <code>/api/pages/payment</code>.
        </p>
      </Container>
    </section>
  )
}

