'use client'

import { useState, useEffect } from 'react'
import Link from 'next/link'
import Image from 'next/image'
import { Container } from '@/components/Container'
import {
  getCart,
  removeFromCart,
  updateCartItemQuantity,
  getCartTotalPrice,
  clearCart,
  type CartItem,
} from '@/lib/cart'
import { api, getPlaceholderImage, getImageUrl, formatPrice } from '@/lib/api'
import styles from './page.module.css'

export default function CartPage() {
  const [cart, setCart] = useState<CartItem[]>([])
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    setCart(getCart())
  }, [])

  const handleRemove = (productId: string) => {
    removeFromCart(productId)
    setCart(getCart())
    window.dispatchEvent(new Event('cartUpdated'))
  }

  const handleQuantityChange = (productId: string, quantity: number) => {
    updateCartItemQuantity(productId, quantity)
    setCart(getCart())
    window.dispatchEvent(new Event('cartUpdated'))
  }

  const total = getCartTotalPrice()

  if (cart.length === 0) {
    return (
      <section className="section">
        <Container size="wide">
          <div className="kicker">корзина</div>
          <h1 className="h2">Корзина пуста</h1>
          <p className="lead">Добавьте товары из каталога</p>
          <Link href="/shop" className={styles.backLink}>
            Перейти в магазин
          </Link>
        </Container>
      </section>
    )
  }

  return (
    <section className="section">
      <Container size="wide">
        <div className="kicker">корзина</div>
        <h1 className="h2">Корзина</h1>

        <div className={styles.cart}>
          <div className={styles.items}>
            {cart.map((item) => (
              <div key={item.productId} className={styles.item}>
                <Link href={`/shop/${item.slug}`} className={styles.imageLink}>
                  <Image
                    src={getImageUrl(item.image)}
                    alt={item.title}
                    width={120}
                    height={120}
                    className={styles.image}
                    unoptimized={item.image?.includes('picsum.photos')}
                  />
                </Link>

                <div className={styles.details}>
                  <Link href={`/shop/${item.slug}`} className={styles.title}>
                    {item.title}
                  </Link>
                  <div className={styles.price}>{formatPrice(item.price, item.currency)}</div>
                </div>

                <div className={styles.quantity}>
                  <button
                    onClick={() => handleQuantityChange(item.productId, item.quantity - 1)}
                    className={styles.quantityButton}
                    disabled={item.quantity <= 1}
                  >
                    −
                  </button>
                  <span className={styles.quantityValue}>{item.quantity}</span>
                  <button
                    onClick={() => handleQuantityChange(item.productId, item.quantity + 1)}
                    className={styles.quantityButton}
                  >
                    +
                  </button>
                </div>

                <div className={styles.itemTotal}>
                  {formatPrice(item.price * item.quantity, item.currency)}
                </div>

                <button
                  onClick={() => handleRemove(item.productId)}
                  className={styles.removeButton}
                  aria-label="Удалить"
                >
                  ×
                </button>
              </div>
            ))}
          </div>

          <div className={styles.summary}>
            <div className={styles.summaryRow}>
              <span>Итого:</span>
              <span className={styles.total}>{formatPrice(total)}</span>
            </div>
            <Link href="/checkout" className={styles.checkoutButton}>
              Оформить заказ
            </Link>
          </div>
        </div>
      </Container>
    </section>
  )
}
