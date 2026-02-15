import { Container } from '@/components/Container'

export default function ReturnsInfoPage() {
  return (
    <section className="section">
      <Container size="wide">
        <div className="kicker">информация</div>
        <h1 className="h2">Возврат</h1>
        <p className="lead">
          TODO: условия возврата. Эндпоинт Go Fiber: <code>/api/pages/returns</code>.
        </p>
      </Container>
    </section>
  )
}

