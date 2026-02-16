'use client'

import { useState, useEffect } from 'react'
import Link from 'next/link'
import { getCartTotalItems } from '@/lib/cart'
import styles from './CartButton.module.css'

export function CartButton() {
  const [itemCount, setItemCount] = useState(0)

  useEffect(() => {
    // Обновляем счётчик при загрузке
    setItemCount(getCartTotalItems())

    // Слушаем изменения localStorage (если корзина меняется в другой вкладке)
    const handleStorageChange = () => {
      setItemCount(getCartTotalItems())
    }
    window.addEventListener('storage', handleStorageChange)

    // Слушаем кастомное событие для обновления в той же вкладке
    const handleCartUpdate = () => {
      setItemCount(getCartTotalItems())
    }
    window.addEventListener('cartUpdated', handleCartUpdate)

    return () => {
      window.removeEventListener('storage', handleStorageChange)
      window.removeEventListener('cartUpdated', handleCartUpdate)
    }
  }, [])

  return (
    <Link href="/cart" className={styles.cartButton}>
      КОРЗИНА
      {itemCount > 0 && <span className={styles.badge}>{itemCount}</span>}
    </Link>
  )
}
