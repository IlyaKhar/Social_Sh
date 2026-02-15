import { Container } from '@/components/Container'

export default function AccountPage() {
  return (
    <section className="section">
      <Container size="wide">
        <div className="kicker">аккаунт</div>
        <h1 className="h2">Личный кабинет</h1>
        <p className="lead">
          TODO: защищённая страница. Здесь будут профиль и заказы, данные берём из эндпоинтов{' '}
          <code>/api/account/me</code> и <code>/api/account/orders</code>.
        </p>
      </Container>
    </section>
  )
}

