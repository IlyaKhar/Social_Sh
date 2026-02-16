'use client'

import { useState, useEffect } from 'react'
import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { CartButton } from './CartButton'
import styles from './PageHeader.module.css'

function isActive(pathname: string, href: string) {
  if (href === '/') return pathname === '/'
  return pathname === href || pathname.startsWith(`${href}/`)
}

export function PageHeader() {
  const pathname = usePathname() ?? '/'
  const [isLoggedIn, setIsLoggedIn] = useState(false)

  useEffect(() => {
    const token = localStorage.getItem('access_token')
    setIsLoggedIn(!!token)
  }, [])

  const shopSectionActive = pathname.startsWith('/shop')
  const gallerySectionActive = pathname.startsWith('/gallery')
  const infoSectionActive = pathname.startsWith('/info') || pathname.startsWith('/contacts')

  return (
    <div className={styles.wrapper}>
      <div className={styles.inner}>
        <Link href="/" className={styles.brandTitle}>
          SOCIAL&nbsp;SH
        </Link>

        <div className={styles.menuBlock}>
          <div className={styles.menuColumn}>
            {/* МАГАЗИН */}
            <Link className={shopSectionActive ? styles.menuPrimaryActive : styles.menuPrimary} href="/shop">
              МАГАЗИН
            </Link>
            {shopSectionActive ? (
              <div className={styles.subList}>
                <Link
                  className={isActive(pathname, '/shop/new') ? styles.subLinkActive : styles.subLink}
                  href="/shop/new"
                >
                  НОВЫЕ ПОСТУПЛЕНИЯ
                </Link>
                <Link
                  className={isActive(pathname, '/shop/sale') ? styles.subLinkActive : styles.subLink}
                  href="/shop/sale"
                >
                  СЕЗОННЫЕ СКИДКИ
                </Link>
                <Link
                  className={isActive(pathname, '/shop') ? styles.subLinkActive : styles.subLink}
                  href="/shop"
                >
                  СМОТРЕТЬ ВСЕ
                </Link>
              </div>
            ) : null}

            {/* ГАЛЕРЕЯ */}
            <Link className={gallerySectionActive ? styles.menuPrimaryActive : styles.menuPrimary} href="/gallery">
              ГАЛЕРЕЯ
            </Link>

            {/* ИНФОРМАЦИЯ */}
            <Link className={infoSectionActive ? styles.menuPrimaryActive : styles.menuPrimary} href="/info/payment">
              ИНФОРМАЦИЯ
            </Link>
            {infoSectionActive ? (
              <div className={styles.subList}>
                <Link
                  className={isActive(pathname, '/info/payment') ? styles.subLinkActive : styles.subLink}
                  href="/info/payment"
                >
                  ОПЛАТА
                </Link>
                <Link
                  className={isActive(pathname, '/info/delivery') ? styles.subLinkActive : styles.subLink}
                  href="/info/delivery"
                >
                  ДОСТАВКА
                </Link>
                <Link
                  className={pathname.startsWith('/contacts') ? styles.subLinkActive : styles.subLink}
                  href="/contacts"
                >
                  КОНТАКТЫ
                </Link>
              </div>
            ) : null}
          </div>
        </div>
      </div>

      <div className={styles.bottomLinks}>
        <CartButton />
        {isLoggedIn ? (
          <Link
            className={isActive(pathname, '/account') ? styles.accountActive : styles.account}
            href="/account"
          >
            ЛИЧНЫЙ КАБИНЕТ
          </Link>
        ) : (
          <Link
            className={isActive(pathname, '/login') ? styles.accountActive : styles.account}
            href="/login"
          >
            ВХОД
          </Link>
        )}
      </div>
    </div>
  )
}


