import Link from 'next/link'
import { Container } from '@/components/Container'
import styles from './Footer.module.css'

export function Footer() {
  return (
    <footer className={styles.root}>
      <Container size="wide">
        <div className={styles.grid}>
          <div className={styles.col}>
            <div className={styles.title}>SOCIAL SH</div>
            <div className={styles.muted}>Ч/б минимал. Быстро. Аккуратно.</div>
          </div>

          <div className={styles.col}>
            <div className={styles.title}>Магазин</div>
            <div className={styles.links}>
              <Link href="/shop/new">Новые поступления</Link>
              <Link href="/shop/sale">Сезонные скидки</Link>
              <Link href="/shop">Смотреть все</Link>
            </div>
          </div>

          <div className={styles.col}>
            <div className={styles.title}>Информация</div>
            <div className={styles.links}>
              <Link href="/info/payment">Оплата</Link>
              <Link href="/info/delivery">Доставка</Link>
              <Link href="/contacts">Контакты</Link>
              <Link href="/info/returns">Возврат</Link>
            </div>
          </div>

          <div className={styles.col}>
            <div className={styles.title}>Правовое</div>
            <div className={styles.links}>
              <Link href="/info/offer">Публичная оферта</Link>
              <Link href="/info/privacy">Персональные данные</Link>
              <Link href="/info/terms">Пользовательское соглашение</Link>
            </div>
          </div>
        </div>

        <div className={styles.bottom}>
          <span className={styles.muted}>© {new Date().getFullYear()} SOCIAL SH</span>
          <span className={styles.muted}>TODO: сюда кинем соцсети/телегу</span>
        </div>
      </Container>
    </footer>
  )
}

