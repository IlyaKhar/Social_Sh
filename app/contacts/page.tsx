import { Container } from '@/components/Container'

export default function ContactsPage() {
  return (
    <section className="section">
      <Container size="wide">
        <div className="kicker">информация</div>
        <h1 className="h2">Контакты</h1>
        <p className="lead">
          TODO: контакты магазина (почта, телефон, соцсети). Данные можно брать из <code>/api/pages/contacts</code>.
        </p>
      </Container>
    </section>
  )
}

