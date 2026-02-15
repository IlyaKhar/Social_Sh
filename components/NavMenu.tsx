'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { useEffect, useMemo, useRef, useState } from 'react'
import { NAV } from '@/components/nav'
import styles from './NavMenu.module.css'

function isActive(pathname: string, href: string) {
  if (href === '/') return pathname === '/'
  return pathname === href || pathname.startsWith(`${href}/`)
}

export function NavMenu() {
  const pathname = usePathname() ?? '/'
  const [isOpen, setIsOpen] = useState(false)
  const dialogRef = useRef<HTMLDivElement | null>(null)

  const groups = useMemo(() => NAV, [])

  useEffect(() => {
    function onKeyDown(e: KeyboardEvent) {
      if (e.key !== 'Escape') return
      setIsOpen(false)
    }

    document.addEventListener('keydown', onKeyDown)
    return () => document.removeEventListener('keydown', onKeyDown)
  }, [])

  useEffect(() => {
    // TODO(фронт): закрывать меню при переходе по роуту (pathname сменился)
    setIsOpen(false)
  }, [pathname])

  return (
    <>
      <div className={styles.desktop}>
        {groups.map((group) => (
          <div key={group.title} className={styles.group}>
            <span className={styles.groupTitle}>{group.title}</span>
            <div className={styles.groupItems}>
              {group.items.map((item) => (
                <Link
                  key={item.href}
                  className={isActive(pathname, item.href) ? styles.linkActive : styles.link}
                  href={item.href}
                >
                  {item.label}
                </Link>
              ))}
            </div>
          </div>
        ))}
      </div>

      <button
        className={styles.burger}
        type="button"
        aria-label="Открыть меню"
        aria-expanded={isOpen}
        onClick={() => setIsOpen((v) => !v)}
      >
        Меню
      </button>

      {isOpen ? (
        <div className={styles.overlay} role="presentation" onClick={() => setIsOpen(false)}>
          <div
            className={styles.dialog}
            role="dialog"
            aria-modal="true"
            aria-label="Меню"
            ref={dialogRef}
            onClick={(e) => e.stopPropagation()}
          >
            <div className={styles.dialogHeader}>
              <span className={styles.dialogTitle}>Навигация</span>
              <button className={styles.close} type="button" onClick={() => setIsOpen(false)}>
                Закрыть
              </button>
            </div>

            <div className={styles.dialogBody}>
              {groups.map((group) => (
                <div key={group.title} className={styles.mGroup}>
                  <div className={styles.mTitle}>{group.title}</div>
                  <div className={styles.mItems}>
                    {group.items.map((item) => (
                      <Link key={item.href} className={styles.mLink} href={item.href}>
                        {item.label}
                      </Link>
                    ))}
                  </div>
                </div>
              ))}
              <div className={styles.mGroup}>
                <div className={styles.mTitle}>Аккаунт</div>
                <div className={styles.mItems}>
                  <Link className={styles.mLink} href="/account">
                    Личный кабинет
                  </Link>
                </div>
              </div>
            </div>
          </div>
        </div>
      ) : null}
    </>
  )
}

