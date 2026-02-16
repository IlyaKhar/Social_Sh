'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { Container } from '@/components/Container'
import { getCart, clearCart, getCartTotalPrice, type CartItem } from '@/lib/cart'
import { api, formatPrice } from '@/lib/api'
import styles from './page.module.css'

export default function CheckoutPage() {
  const router = useRouter()
  const [cart, setCart] = useState<CartItem[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [success, setSuccess] = useState(false)

  const [formData, setFormData] = useState({
    name: '',
    email: '',
    phone: '',
    telegram: '',
    address: '',
    comment: '',
  })

  useEffect(() => {
    const cartItems = getCart()
    if (cartItems.length === 0) {
      router.push('/cart')
      return
    }
    setCart(cartItems)
  }, [router])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)
    setError(null)

    try {
      await api.createOrder({
        items: cart.map((item) => ({
          productId: item.productId,
          quantity: item.quantity,
          price: item.price,
        })),
        customer: {
          name: formData.name,
          email: formData.email,
          phone: formData.phone,
          telegram: formData.telegram,
          address: formData.address,
        },
        comment: formData.comment,
        total: getCartTotalPrice(),
      })

      clearCart()
      setSuccess(true)
      window.dispatchEvent(new Event('cartUpdated'))

      setTimeout(() => {
        router.push('/')
      }, 3000)
    } catch (err: any) {
      setError(err.message || 'Ошибка при оформлении заказа')
    } finally {
      setLoading(false)
    }
  }

  if (success) {
    return (
      <section className="section">
        <Container size="wide">
          <div className="kicker">заказ</div>
          <h1 className="h2">Заказ оформлен!</h1>
          <p className="lead">
            Спасибо за заказ! Мы свяжемся с вами в ближайшее время.
          </p>
        </Container>
      </section>
    )
  }

  const total = getCartTotalPrice()

  return (
    <section className="section">
      <Container size="wide">
        <div className="kicker">оформление заказа</div>
        <h1 className="h2">Оформление заказа</h1>

        <div className={styles.checkout}>
          <form onSubmit={handleSubmit} className={styles.form}>
            <h3>Контактные данные</h3>

            <div className={styles.field}>
              <label htmlFor="name">
                Имя <span className={styles.required}>*</span>
              </label>
              <input
                type="text"
                id="name"
                required
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              />
            </div>

            <div className={styles.field}>
              <label htmlFor="email">
                Email <span className={styles.required}>*</span>
              </label>
              <input
                type="email"
                id="email"
                required
                value={formData.email}
                onChange={(e) => setFormData({ ...formData, email: e.target.value })}
              />
            </div>

            <div className={styles.field}>
              <label htmlFor="phone">Телефон</label>
              <input
                type="tel"
                id="phone"
                value={formData.phone}
                onChange={(e) => setFormData({ ...formData, phone: e.target.value })}
              />
            </div>

            <div className={styles.field}>
              <label htmlFor="telegram">Telegram</label>
              <input
                type="text"
                id="telegram"
                placeholder="@username"
                value={formData.telegram}
                onChange={(e) => setFormData({ ...formData, telegram: e.target.value })}
              />
            </div>

            <div className={styles.field}>
              <label htmlFor="address">Адрес доставки</label>
              <textarea
                id="address"
                rows={3}
                value={formData.address}
                onChange={(e) => setFormData({ ...formData, address: e.target.value })}
              />
            </div>

            <div className={styles.field}>
              <label htmlFor="comment">Комментарий к заказу</label>
              <textarea
                id="comment"
                rows={4}
                value={formData.comment}
                onChange={(e) => setFormData({ ...formData, comment: e.target.value })}
              />
            </div>

            {error && <div className={styles.error}>{error}</div>}

            <button type="submit" disabled={loading} className={styles.submitButton}>
              {loading ? 'Оформление...' : `Оформить заказ на ${formatPrice(total)}`}
            </button>
          </form>

          <div className={styles.summary}>
            <h3>Ваш заказ</h3>
            <div className={styles.items}>
              {cart.map((item) => (
                <div key={item.productId} className={styles.summaryItem}>
                  <span>{item.title}</span>
                  <span>
                    {item.quantity} × {formatPrice(item.price, item.currency)}
                  </span>
                </div>
              ))}
            </div>
            <div className={styles.totalRow}>
              <span>Итого:</span>
              <span className={styles.total}>{formatPrice(total)}</span>
            </div>
          </div>
        </div>
      </Container>
    </section>
  )
}
