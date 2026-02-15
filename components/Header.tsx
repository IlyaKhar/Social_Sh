import Link from 'next/link'
import { NavMenu } from '@/components/NavMenu'
import styles from './Header.module.css'

export function Header() {
  return (
    <header className={styles.root}>
      <div className={styles.inner}>
        <div className={styles.brand}>
          <Link className={styles.logo} href="/">
            SOCIAL&nbsp;SH
          </Link>
          <span className={styles.dot} aria-hidden />
          <span className={styles.tag}>store</span>
        </div>

        <nav className={styles.nav} aria-label="Основная навигация">
          <NavMenu />
        </nav>
      </div>
    </header>
  )
}

