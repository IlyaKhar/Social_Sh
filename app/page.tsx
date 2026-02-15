import Link from 'next/link'
import { Container } from '@/components/Container'
import styles from './page.module.css'

export default function HomePage() {
  return (
    <section className={styles.hero}>
      <Container size="wide">
        <div className={styles.heroGrid}>
          <div className={styles.leftWall} aria-hidden>
            TODO: сюда подвяжем витрину/фото, как сетку товаров у Гоши.
          </div>

          <div className={styles.cards}>
            <Link className={styles.card} href="/shop/new">
              <div className={styles.cardTitle}>Новые поступления</div>
              <div className={styles.cardMeta}>TODO: выводить товары с флагом isNew</div>
            </Link>
            <Link className={styles.card} href="/shop/sale">
              <div className={styles.cardTitle}>Сезонные скидки</div>
              <div className={styles.cardMeta}>TODO: выводить товары с флагом isOnSale</div>
            </Link>
            <Link className={styles.card} href="/projects">
              <div className={styles.cardTitle}>Проекты</div>
              <div className={styles.cardMeta}>TODO: карточки проектов из API</div>
            </Link>
            <Link className={styles.card} href="/info/delivery">
              <div className={styles.cardTitle}>Доставка</div>
              <div className={styles.cardMeta}>Статика/Markdown из API</div>
            </Link>
          </div>
        </div>
      </Container>
    </section>
  )
}

