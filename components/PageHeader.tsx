'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import styles from './PageHeader.module.css'

function isActive(pathname: string, href: string) {
  if (href === '/') return pathname === '/'
  return pathname === href || pathname.startsWith(`${href}/`)
}

export function PageHeader() {
  const pathname = usePathname() ?? '/'

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
            {gallerySectionActive ? (
              <div className={styles.subList}>
                <span className={styles.subLink}>ИНТРО</span>
                <span className={styles.subLink}>ИНТРО 2</span>
                <span className={styles.subLink}>ИНТРО 3</span>
                <span className={styles.subLink}>ТАТУ</span>
                <span className={styles.subLink}>ТОКИО</span>
                <span className={styles.subLink}>ПРОХОР</span>
                <span className={styles.subLink}>НАЗАР</span>
              </div>
            ) : null}

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

      <Link
        className={isActive(pathname, '/account') ? styles.accountActive : styles.account}
        href="/account"
      >
        ЛИЧНЫЙ КАБИНЕТ
      </Link>
    </div>
  )
}


