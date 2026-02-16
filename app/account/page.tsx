'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import Link from 'next/link'
import { Container } from '@/components/Container'
import { api, type User, type Order } from '@/lib/api'

export default function AccountPage() {
  const router = useRouter()
  const [user, setUser] = useState<User | null>(null)
  const [orders, setOrders] = useState<Order[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [isAuthorized, setIsAuthorized] = useState(false)

  useEffect(() => {
    async function loadAccount() {
      const token = localStorage.getItem('access_token')
      if (!token) {
        router.push('/login')
        return
      }

      try {
        const [userData, ordersData] = await Promise.all([
          api.getAccount(),
          api.getOrders().catch(() => ({ items: [] })), // Заказы могут быть пустыми
        ])
        setUser(userData.user)
        setOrders(ordersData.items || [])
        setIsAuthorized(true)
      } catch (e) {
        const errorMsg = e instanceof Error ? e.message : 'Ошибка загрузки данных'
        // Если 401 - токен невалидный, редиректим на логин
        if (errorMsg.includes('401') || errorMsg.includes('Unauthorized')) {
          router.push('/login')
          return
        }
        setError(errorMsg)
        // Не удаляем токен сразу, может быть временная ошибка сети
      } finally {
        setLoading(false)
      }
    }

    loadAccount()
  }, [])

  const handleLogout = () => {
    localStorage.removeItem('access_token')
    router.push('/')
    router.refresh()
  }

  if (loading) {
    return (
      <section className="section">
        <Container size="wide">
          <div className="kicker">аккаунт</div>
          <h1 className="h2">Загрузка...</h1>
        </Container>
      </section>
    )
  }

  if (error || !isAuthorized) {
    return (
      <section className="section">
        <Container size="wide">
          <div className="kicker">аккаунт</div>
          <h1 className="h2">Ошибка</h1>
          <p className="lead">{error || 'Необходима авторизация'}</p>
          <Link href="/login" style={{ marginTop: '1rem', display: 'inline-block' }}>
            Войти
          </Link>
        </Container>
      </section>
    )
  }

  return (
    <section className="section">
      <Container size="wide">
        <div className="kicker">аккаунт</div>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1rem' }}>
          <h1 className="h2">Личный кабинет</h1>
          <button
            onClick={handleLogout}
            style={{
              padding: '0.5rem 1rem',
              background: 'transparent',
              border: '1px solid var(--line)',
              color: 'var(--fg)',
              fontFamily: 'var(--font-mono)',
              cursor: 'pointer',
            }}
          >
            Выйти
          </button>
        </div>

        <div style={{ marginTop: '2rem' }}>
          <h2 style={{ fontSize: '1.25rem', marginBottom: '1rem' }}>Профиль</h2>
          <div style={{ padding: '1.5rem', border: '1px solid var(--line)', marginBottom: '2rem' }}>
            <p><strong>Имя:</strong> {user?.name || 'Не указано'}</p>
            <p><strong>Email:</strong> {user?.email}</p>
            <p><strong>Роль:</strong> {user?.role}</p>
          </div>

          <h2 style={{ fontSize: '1.25rem', marginBottom: '1rem' }}>Заказы ({orders.length})</h2>
          {orders.length === 0 ? (
            <p className="lead">У вас пока нет заказов</p>
          ) : (
            <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
              {orders.map((order) => (
                <div key={order.id} style={{ padding: '1.5rem', border: '1px solid var(--line)' }}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '1rem' }}>
                    <div>
                      <strong>Заказ #{order.id.slice(0, 8)}</strong>
                      <div style={{ fontSize: '0.875rem', color: 'var(--muted)', marginTop: '0.25rem' }}>
                        {new Date(order.createdAt).toLocaleDateString('ru-RU')}
                      </div>
                    </div>
                    <div>
                      <div><strong>{order.total / 100} ₽</strong></div>
                      <div style={{ fontSize: '0.875rem', color: 'var(--muted)' }}>{order.status}</div>
                    </div>
                  </div>
                  {order.items && order.items.length > 0 && (
                    <div style={{ marginTop: '1rem', paddingTop: '1rem', borderTop: '1px solid var(--line)' }}>
                      {order.items.map((item) => (
                        <div key={item.id} style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '0.5rem' }}>
                          <span>{item.title} × {item.quantity}</span>
                          <span>{item.price / 100} ₽</span>
                        </div>
                      ))}
                    </div>
                  )}
                </div>
              ))}
            </div>
          )}
        </div>
      </Container>
    </section>
  )
}
