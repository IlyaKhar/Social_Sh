'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import Link from 'next/link'
import { Container } from '@/components/Container'
import { api } from '@/lib/api'
import styles from './page.module.css'

export default function SignupPage() {
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
    const name = formData.get('name') as string

    if (password.length < 8) {
      setError('Пароль должен быть не менее 8 символов')
      setLoading(false)
      return
    }

    try {
      const data = await api.signUp(email, password, name)
      localStorage.setItem('access_token', data.access)
      router.push('/account')
      router.refresh()
    } catch (err: any) {
      const errorMsg = err?.message || 'Ошибка регистрации'
      // Обрабатываем разные типы ошибок
      if (errorMsg.includes('409') || errorMsg.includes('уже существует')) {
        setError('Пользователь с таким email уже зарегистрирован')
      } else if (errorMsg.includes('404') || errorMsg.includes('Not Found')) {
        setError('Сервер не отвечает. Проверьте что бэкенд запущен на порту 3001')
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
        <div className="kicker">регистрация</div>
        <h1 className="h2">Регистрация</h1>

        <form onSubmit={handleSubmit} className={styles.form}>
          <div className={styles.field}>
            <label htmlFor="name">Имя</label>
            <input
              type="text"
              id="name"
              name="name"
              required
              placeholder="Ваше имя"
            />
          </div>

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
              minLength={8}
              placeholder="Минимум 8 символов"
              autoComplete="new-password"
            />
          </div>

          {error && <div className={styles.error}>{error}</div>}

          <button type="submit" disabled={loading} className={styles.submit}>
            {loading ? 'Регистрация...' : 'Зарегистрироваться'}
          </button>

          <div className={styles.links}>
            <p>
              Уже есть аккаунт? <Link href="/login">Войти</Link>
            </p>
          </div>
        </form>
      </Container>
    </section>
  )
}
