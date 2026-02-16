import Link from 'next/link'
import { Container } from '@/components/Container'
import { ProductGrid } from '@/components/ProductGrid'
import { api } from '@/lib/api'

// Отключаем кеширование - всегда загружаем свежие данные с бэка
export const revalidate = 0

export default async function ShopNewPage() {
  let products: any[] = []
  let error: string | null = null

  try {
    const response = await api.getProducts({ new: true, limit: 50 })
    products = response.items || []
  } catch (e) {
    error = e instanceof Error ? e.message : 'Ошибка загрузки товаров'
    console.error('Failed to load products:', e)
  }

  return (
    <section className="section">
      <Container size="wide">
        <div className="kicker">магазин</div>
        <h1 className="h2">Новые поступления</h1>

        <div style={{ marginTop: '1rem', display: 'flex', gap: '0.75rem', flexWrap: 'wrap', marginBottom: '2rem' }}>
          <Link href="/shop">Смотреть все</Link>
          <Link href="/shop/sale">Сезонные скидки</Link>
        </div>

        {error ? (
          <div style={{ color: 'var(--muted)', padding: '2rem 0' }}>{error}</div>
        ) : (
          <ProductGrid products={products} />
        )}
      </Container>
    </section>
  )
}
