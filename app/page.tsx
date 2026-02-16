import Link from 'next/link'
import { Container } from '@/components/Container'
import { ProductGrid } from '@/components/ProductGrid'
import { api } from '@/lib/api'
import styles from './page.module.css'

// Отключаем кеширование - всегда загружаем свежие данные с бэка
export const revalidate = 0

export default async function HomePage() {
  let newProducts: any[] = []
  let saleProducts: any[] = []

  try {
    const [newResponse, saleResponse] = await Promise.all([
      api.getProducts({ new: true, limit: 4 }).catch(() => ({ items: [] })),
      api.getProducts({ sale: true, limit: 4 }).catch(() => ({ items: [] })),
    ])
    newProducts = newResponse.items || []
    saleProducts = saleResponse.items || []
  } catch (e) {
    console.error('Failed to load products:', e)
  }

  return (
    <section className={styles.hero}>
      <Container size="wide">
        <div className={styles.heroGrid}>
          <div className={styles.leftWall} aria-hidden>
            {newProducts.length > 0 && (
              <div style={{ display: 'grid', gap: '1rem' }}>
                <ProductGrid products={newProducts.slice(0, 3)} />
              </div>
            )}
          </div>

          <div className={styles.cards}>
            <Link className={styles.card} href="/shop/new">
              <div className={styles.cardTitle}>Новые поступления</div>
              <div className={styles.cardMeta}>
                {newProducts.length > 0 ? `${newProducts.length} товаров` : 'Скоро'}
              </div>
            </Link>
            <Link className={styles.card} href="/shop/sale">
              <div className={styles.cardTitle}>Сезонные скидки</div>
              <div className={styles.cardMeta}>
                {saleProducts.length > 0 ? `${saleProducts.length} товаров` : 'Скоро'}
              </div>
            </Link>
            <Link className={styles.card} href="/gallery">
              <div className={styles.cardTitle}>Галерея</div>
              <div className={styles.cardMeta}>Фотографии</div>
            </Link>
            <Link className={styles.card} href="/info/delivery">
              <div className={styles.cardTitle}>Доставка</div>
              <div className={styles.cardMeta}>Информация</div>
            </Link>
          </div>
        </div>
      </Container>
    </section>
  )
}
