'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import Link from 'next/link'
import { Container } from '@/components/Container'
import { api } from '@/lib/api'
import styles from './page.module.css'

export default function LoginPage() {
  const router = useRouter()
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    setLoading(true)
    setError(null)

    const formData = new FormData(e.currentTarget)
    const email = formData.get('email') as string
    const password = formData.get('password') as string

    try {
      const data = await api.signIn(email, password)
      localStorage.setItem('access_token', data.access)
      router.push('/account')
      router.refresh()
    } catch (err: any) {
      const errorMsg = err?.message || 'Ошибка входа'
      if (errorMsg.includes('404') || errorMsg.includes('Not Found')) {
        setError('Сервер не отвечает. Проверьте что бэкенд запущен на порту 3001')
      } else if (errorMsg.includes('401') || errorMsg.includes('неверный')) {
        setError('Неверный email или пароль')
      } else {
        setError(errorMsg)
      }
    } finally {
      setLoading(false)
    }
  }

  return (
    <section className="section">
      <Container size="wide">
        <div className="kicker">авторизация</div>
        <h1 className="h2">Вход</h1>

        <form onSubmit={handleSubmit} className={styles.form}>
          <div className={styles.field}>
            <label htmlFor="email">Email</label>
            <input
              type="email"
              id="email"
              name="email"
              required
              placeholder="your@email.com"
            />
          </div>

          <div className={styles.field}>
            <label htmlFor="password">Пароль</label>
            <input
              type="password"
              id="password"
              name="password"
              required
              placeholder="••••••••"
            />
          </div>

          {error && <div className={styles.error}>{error}</div>}

          <button type="submit" disabled={loading} className={styles.submit}>
            {loading ? 'Вход...' : 'Войти'}
          </button>

          <div className={styles.links}>
            <p>
              Нет аккаунта? <Link href="/signup">Зарегистрироваться</Link>
            </p>
          </div>
        </form>
      </Container>
    </section>
  )
}
